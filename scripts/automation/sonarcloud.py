#!/usr/bin/env python3
"""SonarCloud management automation for terraform-provider-cvp.

SonarCloud has no dedicated management CLI (`sonar-scanner` only *runs* scans),
so project administration goes through the Web API. This helper wraps the calls
this project needs, authenticating with a personal management token.

Token: export SONAR_MGMT_TOKEN, e.g.

    export SONAR_MGMT_TOKEN="$(gopass show -o infra4/sonarcloud/management-token)"

Usage:
    sonarcloud.py status                       Show the quality-gate status.
    sonarcloud.py gate                          Show the assigned quality gate.
    sonarcloud.py new-code                       Show the New Code definition.
    sonarcloud.py set-new-code previous_version  Set New Code to the previous version.
    sonarcloud.py set-new-code days 30           Set New Code to a rolling 30 days.
"""

from __future__ import annotations

import argparse
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
from base64 import b64encode

BASE_URL = "https://sonarcloud.io"
ORGANIZATION = "ioplane"
PROJECT_KEY = "ioplane_terraform-provider-cvp"


def _token() -> str:
    token = os.environ.get("SONAR_MGMT_TOKEN")
    if not token:
        print(
            "SONAR_MGMT_TOKEN is not set; export it from gopass:\n"
            '  export SONAR_MGMT_TOKEN="$(gopass show -o infra4/sonarcloud/management-token)"',
            file=sys.stderr,
        )
        raise SystemExit(2)
    return token


def _call(method: str, path: str, params: dict[str, str] | None = None) -> dict:
    query = urllib.parse.urlencode(params or {})
    url = f"{BASE_URL}{path}"
    data = None
    if method == "POST":
        data = query.encode()
    elif query:
        url = f"{url}?{query}"
    if not url.startswith("https://"):  # defence-in-depth for urlopen
        msg = "refusing non-HTTPS SonarCloud URL"
        raise ValueError(msg)
    auth = b64encode(f"{_token()}:".encode()).decode()
    request = urllib.request.Request(url, data=data, method=method)  # noqa: S310 — https-only, asserted above
    request.add_header("Authorization", f"Basic {auth}")
    try:
        with urllib.request.urlopen(request, timeout=30) as response:  # noqa: S310
            body = response.read().decode()
    except urllib.error.HTTPError as exc:
        print(f"SonarCloud API {exc.code}: {exc.read().decode()[:300]}", file=sys.stderr)
        raise SystemExit(1) from exc
    return json.loads(body) if body else {}


def cmd_status(_: argparse.Namespace) -> int:
    result = _call("GET", "/api/qualitygates/project_status", {"projectKey": PROJECT_KEY})
    status = result.get("projectStatus", {})
    print(f"quality gate: {status.get('status', 'UNKNOWN')}")
    for condition in status.get("conditions", []):
        print(
            f"  {condition.get('status'):5} {condition.get('metricKey')}: "
            f"{condition.get('actualValue')} (threshold {condition.get('errorThreshold')})"
        )
    return 0


def cmd_gate(_: argparse.Namespace) -> int:
    result = _call(
        "GET",
        "/api/qualitygates/get_by_project",
        {"organization": ORGANIZATION, "project": PROJECT_KEY},
    )
    gate = result.get("qualityGate", {})
    print(f"quality gate: {gate.get('name', '?')} (default={gate.get('default', False)})")
    return 0


def cmd_new_code(_: argparse.Namespace) -> int:
    result = _call("GET", "/api/new_code_periods/list", {"project": PROJECT_KEY})
    for period in result.get("newCodePeriods", []) or [result]:
        branch = period.get("branchKey", "main")
        print(f"branch={branch} type={period.get('type')} value={period.get('value', '')}")
    return 0


def cmd_set_new_code(ns: argparse.Namespace) -> int:
    params = {"project": PROJECT_KEY, "type": ns.type.upper()}
    if ns.value:
        params["value"] = ns.value
    _call("POST", "/api/new_code_periods/set", params)
    print(f"New Code set: type={ns.type} value={ns.value or ''}")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(prog="sonarcloud.py", description=__doc__)
    sub = parser.add_subparsers(dest="cmd", required=True)
    sub.add_parser("status", help="Show the quality-gate status.").set_defaults(func=cmd_status)
    sub.add_parser("gate", help="Show the assigned quality gate.").set_defaults(func=cmd_gate)
    sub.add_parser("new-code", help="Show the New Code definition.").set_defaults(func=cmd_new_code)
    p_set = sub.add_parser("set-new-code", help="Set the New Code definition.")
    p_set.add_argument("type", choices=["previous_version", "days", "specific_analysis"])
    p_set.add_argument("value", nargs="?", default="")
    p_set.set_defaults(func=cmd_set_new_code)

    ns = parser.parse_args()
    return ns.func(ns)


if __name__ == "__main__":
    raise SystemExit(main())

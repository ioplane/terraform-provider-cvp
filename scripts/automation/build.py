#!/usr/bin/env python3
"""podman-py automation for terraform-provider-cvp.

A thin, dependency-light wrapper over the same dev container the Makefile and
podman-compose stacks drive. Exists so CI or ad-hoc scripts can manage the
lifecycle without shelling out through Make.

Docs:
  - podman-py: https://github.com/containers/podman-py
  - Compose Specification: https://github.com/compose-spec/compose-spec

Usage:
  scripts/automation/build.py up                 # build image + start dev container
  scripts/automation/build.py exec -- go test ./...
  scripts/automation/build.py down
  scripts/automation/build.py versions
"""

from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
COMPOSE_FILE = REPO_ROOT / "deployments" / "compose" / "compose.dev.yml"
SERVICE = "dev"


def compose(*args: str) -> int:
    """Run podman-compose against the dev stack."""
    cmd = ["podman-compose", "-f", str(COMPOSE_FILE), *args]
    return subprocess.call(cmd)


def cmd_up(_: argparse.Namespace) -> int:
    return compose("up", "-d", "--build")


def cmd_down(_: argparse.Namespace) -> int:
    return compose("down")


def cmd_exec(ns: argparse.Namespace) -> int:
    if not ns.command:
        print("error: nothing to exec; pass a command after --", file=sys.stderr)
        return 2
    return compose("exec", "-T", SERVICE, *ns.command)


def cmd_versions(_: argparse.Namespace) -> int:
    script = (
        "go version && terraform version | head -1 && "
        "golangci-lint version --short && tfplugindocs --version 2>/dev/null || true"
    )
    return compose("exec", "-T", SERVICE, "bash", "-lc", script)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="cmd", required=True)

    sub.add_parser("up", help="build image + start dev container").set_defaults(func=cmd_up)
    sub.add_parser("down", help="stop + remove dev container").set_defaults(func=cmd_down)
    sub.add_parser("versions", help="print toolchain versions").set_defaults(func=cmd_versions)

    exec_p = sub.add_parser("exec", help="run a command in the dev container")
    exec_p.add_argument("command", nargs=argparse.REMAINDER, help="command after --")
    exec_p.set_defaults(func=cmd_exec)

    ns = parser.parse_args()
    return ns.func(ns)


if __name__ == "__main__":
    raise SystemExit(main())

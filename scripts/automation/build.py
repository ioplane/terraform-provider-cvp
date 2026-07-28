#!/usr/bin/env python3
"""Build automation for terraform-provider-cvp via podman-py.

Drives the dev-image build and container lifecycle through the Podman REST
socket (not shelling out to the CLI), so CI and ad-hoc scripts can manage the
same container the Taskfile and podman-compose stacks use.

Spec references:
  - Containerfile.5  https://github.com/containers/common/blob/main/docs/Containerfile.5.md
  - Compose Spec     https://github.com/compose-spec/compose-spec/blob/main/spec.md
  - containers.conf  https://github.com/containers/common/blob/main/docs/containers.conf.5.md
  - podman-py        https://github.com/containers/podman-py

Requires:
  - python >= 3.11
  - podman-py >= 5.7

Usage:
  scripts/automation/build.py dev-image      Build the development image.
  scripts/automation/build.py up             Start the dev container.
  scripts/automation/build.py down           Stop and remove the dev container.
  scripts/automation/build.py exec -- <cmd>  Run <cmd> inside the dev container.
  scripts/automation/build.py lint           Run all lint targets in one shot.
  scripts/automation/build.py versions       Print the toolchain versions.
"""

from __future__ import annotations

import argparse
import os
import sys
from dataclasses import dataclass
from pathlib import Path

try:
    from podman import PodmanClient
    from podman.errors import APIError, NotFound
except ImportError as exc:  # pragma: no cover
    print(f"podman-py not installed: {exc}", file=sys.stderr)
    sys.exit(2)

REPO_ROOT = Path(__file__).resolve().parents[2]
DEV_IMAGE = "localhost/terraform-provider-cvp-dev:local"
DEV_CONTAINER = "terraform-provider-cvp-dev"
DEV_CONTAINERFILE = REPO_ROOT / "deployments" / "containers" / "Containerfile.dev"

_DEV_ENVIRONMENT = {
    "GOFLAGS": "-buildvcs=false",
    "GOCACHE": "/tmp/go-cache",  # noqa: S108 — container path, matches compose.dev.yml
    "GOMODCACHE": "/go/pkg/mod",
    "CHECKPOINT_DISABLE": "1",
    "TF_IN_AUTOMATION": "1",
}

# Lab credentials the caller may export at any time; forwarded into every exec
# so acceptance runs authenticate even against a container started earlier.
_LAB_ENV_VARS = ("CVP_ENDPOINT", "CVP_AUTH_METHOD", "CVP_TOKEN", "TF_ACC")


def _lab_environment() -> dict[str, str]:
    return {var: value for var in _LAB_ENV_VARS if (value := os.environ.get(var))}


@dataclass(frozen=True, slots=True)
class Settings:
    """Runtime configuration derived from the environment."""

    socket_uri: str

    @classmethod
    def from_env(cls) -> Settings:
        sock = os.environ.get("CONTAINER_HOST") or os.environ.get("PODMAN_SOCKET")
        if not sock:
            uid = os.geteuid()
            sock = f"unix:///run/user/{uid}/podman/podman.sock" if uid else "unix:///run/podman/podman.sock"
        return cls(socket_uri=sock)


def client(settings: Settings) -> PodmanClient:
    return PodmanClient(base_url=settings.socket_uri)


# --- commands ---------------------------------------------------------------


def cmd_dev_image(settings: Settings) -> int:
    with client(settings) as podman:
        print(f"build: {DEV_CONTAINERFILE} -> {DEV_IMAGE}")
        try:
            _, stream = podman.images.build(
                path=str(REPO_ROOT),
                dockerfile=str(DEV_CONTAINERFILE.relative_to(REPO_ROOT)),
                tag=DEV_IMAGE,
                rm=True,
                forcerm=True,
            )
        except APIError as exc:
            print(f"podman build failed: {exc}", file=sys.stderr)
            return 1
        for chunk in stream:
            msg = chunk.get("stream") or chunk.get("error") or "" if isinstance(chunk, dict) else str(chunk)
            if msg:
                sys.stdout.write(msg)
        return 0


def cmd_up(settings: Settings) -> int:
    with client(settings) as podman:
        try:
            existing = podman.containers.get(DEV_CONTAINER)
        except NotFound:
            existing = None
        if existing is not None:
            if existing.status != "running":
                existing.start()
            print(f"{DEV_CONTAINER} running")
            return 0

        if not podman.images.exists(DEV_IMAGE):
            rc = cmd_dev_image(settings)
            if rc:
                return rc

        container = podman.containers.create(
            image=DEV_IMAGE,
            name=DEV_CONTAINER,
            command=["sleep", "infinity"],
            working_dir="/app",
            mounts=[{"type": "bind", "source": str(REPO_ROOT), "target": "/app", "read_only": False}],
            environment=dict(_DEV_ENVIRONMENT),
            cap_drop=["ALL"],
            security_opt=["no-new-privileges:true"],
            restart_policy={"Name": "unless-stopped"},
        )
        container.start()
        print(f"{DEV_CONTAINER} started ({container.id[:12]})")
        return 0


def cmd_down(settings: Settings) -> int:
    with client(settings) as podman:
        try:
            container = podman.containers.get(DEV_CONTAINER)
        except NotFound:
            print(f"{DEV_CONTAINER}: not present")
            return 0
        container.stop()
        container.remove(force=True)
        print(f"{DEV_CONTAINER}: stopped and removed")
        return 0


def cmd_exec(settings: Settings, command: list[str]) -> int:
    if not command:
        print("usage: build.py exec -- <command...>", file=sys.stderr)
        return 2
    with client(settings) as podman:
        try:
            container = podman.containers.get(DEV_CONTAINER)
        except NotFound:
            print(f"{DEV_CONTAINER} not running; run `build.py up` first", file=sys.stderr)
            return 1
        rc, output = container.exec_run(cmd=command, demux=False, tty=True, environment=_lab_environment())
        if isinstance(output, (bytes, bytearray)):
            sys.stdout.buffer.write(output)
        elif isinstance(output, str):
            sys.stdout.write(output)
        return int(rc or 0)


def cmd_lint(settings: Settings) -> int:
    targets = [
        ["golangci-lint", "run", "./..."],
        ["markdownlint-cli2", "**/*.md"],
        ["yamllint", "-c", ".yamllint.yaml", "."],
        ["cspell", "--no-progress", "--no-summary", "--config", ".cspell.json", "**/*.md", "**/*.go"],
        ["govulncheck", "./..."],
    ]
    rc = 0
    for cmd in targets:
        print(f">>> {' '.join(cmd)}")
        rc = max(rc, cmd_exec(settings, cmd))
    return rc


def cmd_versions(settings: Settings) -> int:
    return cmd_exec(
        settings,
        ["bash", "-lc", "go version && terraform version | head -1 && golangci-lint version --short"],
    )


# --- entry point ------------------------------------------------------------


def main(argv: list[str]) -> int:
    parser = argparse.ArgumentParser(prog="build.py", add_help=True)
    sub = parser.add_subparsers(dest="cmd", required=True)
    sub.add_parser("dev-image", help="Build the development image.")
    sub.add_parser("up", help="Start the development container.")
    sub.add_parser("down", help="Stop and remove the development container.")
    p_exec = sub.add_parser("exec", help="Run a command inside the dev container.")
    p_exec.add_argument("command", nargs=argparse.REMAINDER)
    sub.add_parser("lint", help="Run all linters one after another.")
    sub.add_parser("versions", help="Print the toolchain versions.")

    args = parser.parse_args(argv)
    settings = Settings.from_env()

    match args.cmd:
        case "dev-image":
            return cmd_dev_image(settings)
        case "up":
            return cmd_up(settings)
        case "down":
            return cmd_down(settings)
        case "exec":
            return cmd_exec(settings, [arg for arg in args.command if arg != "--"])
        case "lint":
            return cmd_lint(settings)
        case "versions":
            return cmd_versions(settings)
        case _:  # pragma: no cover
            parser.error(f"unknown command: {args.cmd}")
            return 2


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))

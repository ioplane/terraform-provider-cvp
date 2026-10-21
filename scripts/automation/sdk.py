"""Prepare the exact unpublished SDK source workspace; never publish local replaces."""

from __future__ import annotations

import argparse
import json
import os
import re
import shutil
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
MODULE = "github.com/ioplane/cvp-sdk/go"


def run(*args: str, **kwargs) -> str:
    return subprocess.check_output(args, cwd=ROOT, text=True, **kwargs).strip()  # noqa: S603 -- fixed git/go argv, no shell.


def prepare() -> Path:
    pin = json.loads((ROOT / ".sdk-source.json").read_text())
    if pin["repository"] != "ioplane/arista-cvp-re" or not re.fullmatch(r"[0-9a-f]{40}", pin["revision"]):
        raise ValueError("SDK source needs the approved repository and full commit SHA")
    source = Path(os.environ.get("CVP_SDK_SOURCE", ROOT / ".cache/cvp-sdk/source")).resolve()
    if run("git", "-C", str(source), "rev-parse", "HEAD") != pin["revision"]:
        raise ValueError("SDK source revision differs from .sdk-source.json")
    if run("git", "-C", str(source), "status", "--porcelain", "--", "sdk/go"):
        raise ValueError("SDK module source must be clean")
    module = source / "sdk/go"
    if (module / "go.mod").read_text().splitlines()[0] != "module " + MODULE:
        raise ValueError("unexpected SDK module identity")
    workspace = ROOT / ".cache/cvp-sdk/go.work"
    workspace.parent.mkdir(parents=True, exist_ok=True)
    version = (ROOT / "go.mod").read_text().split("\ngo ", 1)[1].splitlines()[0]
    workspace.write_text(f"go {version}\n\nuse (\n{json.dumps(str(ROOT))}\n{json.dumps(str(module))}\n)\n")
    return module


def tidy(*, check: bool) -> None:
    source = prepare()
    with tempfile.TemporaryDirectory(dir=ROOT / ".cache/cvp-sdk") as directory:
        temporary = Path(directory)
        for name in ("go.mod", "go.sum"):
            shutil.copy2(ROOT / name, temporary / name)
        modfile = "-modfile=" + str(temporary / "go.mod")
        env = {**os.environ, "GOWORK": "off"}
        run(
            "go",
            "mod",
            "edit",
            modfile,
            "-require=" + MODULE + "@v0.0.0",
            "-replace=" + MODULE + "=" + str(source),
            env=env,
        )
        run("go", "mod", "tidy", modfile, env=env)
        run("go", "mod", "edit", modfile, "-droprequire=" + MODULE, "-dropreplace=" + MODULE, env=env)
        for name in ("go.mod", "go.sum"):
            desired = (temporary / name).read_bytes()
            if check:
                if (ROOT / name).read_bytes() != desired:
                    raise ValueError(f"{name} differs from tidy; run task tidy")
            else:
                (ROOT / name).write_bytes(desired)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("command", choices=("prepare", "tidy", "check-tidy"))
    command = parser.parse_args().command
    if command == "prepare":
        prepare()
    else:
        tidy(check=command == "check-tidy")

#!/usr/bin/env bash
# PreToolUse(Bash) guard: nudge toward source-of-truth when a command hand-rolls
# a CVP/EOS API call (raw curl to a CVP endpoint) without going through cvprac
# or having consulted the proto/re-docs. Non-blocking — emits a reminder only.
# See AGENTS.md "Source of truth — never from memory" and the cvp-api-lookup skill.
python3 - <<'PY'
import json, sys, re
try:
    data = json.load(sys.stdin)
except Exception:
    sys.exit(0)
cmd = (data.get("tool_input", {}) or {}).get("command", "") or ""

# Only care about raw HTTP calls to CVP-ish endpoints.
hits_cvp = re.search(r"cvpservice|/api/resources/|/api/v\d|command-api|createEnrollmentToken|/enroll|ingestgrpcurl|cvaddr|TerminAttr", cmd)
uses_source = re.search(r"cvprac|cvp_api\.py|arista-cvp-re|\.proto|arista-mcp", cmd)
raw_http = re.search(r"\bcurl\b|\bwget\b|http[s]?://[^ ]*cvp", cmd)

if hits_cvp and raw_http and not uses_source:
    msg = ("CVP/EOS API guard: this looks like a hand-rolled CVP call. Before "
           "trusting the endpoint/fields, confirm them from source — read the "
           "matching method in ../cvprac/cvprac/cvp_api.py (current, "
           "apiversion-gated endpoint) or the cloudvision-apis proto, and prefer "
           "running cvprac over raw curl. A 4xx from a guessed path is NOT "
           "evidence about account permissions. (skill: cvp-api-lookup)")
    print(json.dumps({
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "additionalContext": msg
        }
    }))
sys.exit(0)
PY

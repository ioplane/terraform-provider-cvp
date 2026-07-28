---
name: cvp-api-lookup
description: Use BEFORE calling or writing any Arista CVP/EOS API — a REST endpoint, gRPC RPC/field, TerminAttr flag, or EOS CLI. Resolves the exact, current API from authoritative sources (cvprac, cloudvision-apis protos, arista-mcp, arista-cvp-re) instead of guessing from memory.
---

# CVP / EOS API lookup — resolve from source, never memory

Guessing a CVP endpoint or field from training recall wastes cycles and produces
false conclusions (e.g. a 403 from a *deprecated* endpoint misread as "account
unauthorized"). Before you build any CVP/EOS request or command, resolve the
exact, current API from a source.

## Procedure (do this first, every time)

1. **Name the fact** you need: REST path + body, gRPC service/RPC/message field,
   TerminAttr flag, EOS CLI, port, or auth/RBAC behavior.
2. **Resolve it** from the highest-preference source that covers it:
   - **REST endpoint / body** → read the matching method in
     `../cvprac/cvprac/cvp_api.py` (search for the verb, e.g. `enroll`, `device`,
     `configlet`, `change_control`, `tag`, `workspace`). Note the
     `apiversion`-gated branch that matches the target CVP version — endpoints
     changed across releases (resource API vs `/cvpservice/...`).
   - **gRPC service / RPC / message field** → read the `cloudvision-apis` proto
     (`api/proto/arista/<svc>.v1/*.proto` if vendored, else fetch the file at the
     pinned `.upstream-revision`). Copy field names from the proto.
   - **Behavior / feature / default / version note** → `arista-mcp` MCP
     (`search_docs`), and cite it.
   - **Why an API is gated / how auth works** → `../arista-cvp-re/docs`
     (ext_authz, certhdr, ingest 9910/9911, onboarding-token model, RBAC).
   - **Any other library/CLI** (Terraform framework, Proxmox `qm`, …) →
     `context7` MCP (`resolve-library-id` → `query-docs`).
3. **Only then** build the call. **Cite** the source (file:line or MCP doc) in
   the commit/PR body or the code comment.
4. If a call fails, **re-check the source before concluding** — especially before
   reading a 4xx as an authorization/permission fact. A 403/404 from the wrong
   (old) path is not evidence about the account's rights.

## Red flags — you are guessing, STOP and look it up

| Thought | Do instead |
|---|---|
| "The endpoint is probably `/cvpservice/...`" | grep `cvprac/cvp_api.py` for the verb |
| "The field is likely `serialNum`/`hostnameOrIp`…" | read the proto |
| "403 → the account is unauthorized" | confirm you used the **current** endpoint (cvprac) |
| "CVP does X by default" | confirm via `arista-mcp` |
| "TerminAttr flag is `-ingestgrpcurl`…" | check arista-mcp / cvprac; flags are version-specific |

## Prefer cvprac for the call itself

When you need to *make* a CVP REST call, prefer running `cvprac`
(`CvpClient().api.<method>()`) over hand-rolled `curl` — it already encodes the
current, version-gated endpoint and auth. Reach for raw `curl` only to probe
something cvprac does not cover, and only after reading the proto/re-docs.

# Submission Packet — Permify/permify-cli Issue #2

Issue: https://github.com/Permify/permify-cli/issues/2
Algora org page: https://algora.io/Permify

## Goal
Implement persistent storage (via the existing CLI config YAML) for:
- endpoint
- bearer token (optional)
- TLS client cert path (optional)
- TLS client cert key path (optional)

Then ensure the permify client uses the stored values during client creation.

## Implementation summary
- Extend `core/config.CoreConfig` to include `token`, `cert_path`, `cert_key` fields in YAML.
- Update `permctl configure` flow to prompt for token/cert paths before listing tenants, store these values to config, then persist to YAML.
- Update `core/client.New` to:
  - use TLS when endpoint is `https://`
  - optionally attach `authorization: Bearer <token>` via `grpc.WithPerRPCCredentials` (TLS-only)
  - optionally present a client cert (mTLS) when both `cert_path` and `cert_key` are provided

## Patch
Local patch file (generated from branch `feat/bounty-2-credential-storage`):
- `Permify-permify-cli-2-credential-storage.patch`

## How to submit (manual)
1) In `/Users/mavenai/Desktop/Singaw Sity Labs/Singaw Sity Codex/bounty-work/permify-cli`, push branch `feat/bounty-2-credential-storage` to your fork.
2) Open a PR to `Permify/permify-cli` targeting `main`.
3) Post a concise public comment on the issue linking the PR (and `/attempt` if Algora requires it).

## Notes / safety
- Token is only sent over TLS; if endpoint is not `https://`, client creation errors.
- No credentials are embedded in code; this only changes config schema + client wiring.

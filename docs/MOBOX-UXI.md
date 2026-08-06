# MoBox UXI

MoBox UXI is the conversation-first operator surface for AIFT-OS. The first vertical slice is deliberately read-only: it persists sessions, discovers local repositories from filesystem evidence, connects to an OpenAI-compatible local model, and refuses to pretend an unavailable model or unexecuted action succeeded.

## Run

```bash
go run ./cmd/aiftd
```

Open `http://127.0.0.1:8787`.

Configuration:

- `AIFT_ROOT`: federation workspace root. Defaults to `~/AIFT` when present.
- `AIFT_UXI_HOME`: persistent UXI state. Defaults to `~/.aift/uxi`.
- `AIFT_UXI_ADDR`: loopback address. Defaults to `127.0.0.1:8787`.
- `AIFT_MODEL_URL`: local OpenAI-compatible endpoint. Defaults to `http://127.0.0.1:8080/v1`.
- `AIFT_MODEL_NAME`: model identifier. Defaults to `local`.

## Truth contract

1. Observed evidence outranks generated claims.
2. AI output is a proposal or explanation, not proof of execution.
3. The first release exposes read-only repository inspection only.
4. Model failure is shown as degraded mode rather than hidden.
5. The daemon refuses non-loopback binding.

## API

- `GET /health`
- `GET /v1/system`
- `GET /v1/repositories`
- `GET /v1/sessions`
- `POST /v1/sessions`
- `GET /v1/sessions/{id}`
- `POST /v1/sessions/{id}/messages`
- `GET /v1/events`

## Next extraction

The next phase should add typed plan, approval, action, invocation, job, artifact, and validation endpoints before enabling any mutation. AIFT-Forge should be integrated through an adapter rather than invoked directly by the browser.

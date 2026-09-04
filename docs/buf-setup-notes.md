# Protobuf / Buf Setup Notes

## What's set up

Project: `goBase` (`~/projects/goBase/pkg/api/v1`)

**`buf.yaml`** — defines the local module and remote dependencies:
```yaml
version: v2
modules:
  - path: .
deps:
  - buf.build/googleapis/googleapis
  - buf.build/grpc-ecosystem/grpc-gateway
lint:
  use:
    - STANDARD
breaking:
  use:
    - FILE
```

**`buf.gen.yaml`** — defines code generation (remote plugins, no local installs needed):
```yaml
version: v2
plugins:
  - remote: buf.build/protocolbuffers/go
    out: gen/go
    opt: paths=source_relative
  - remote: buf.build/grpc-ecosystem/gateway
    out: gen/go
    opt: paths=source_relative
  - remote: buf.build/grpc-ecosystem/openapiv2
    out: gen/openapiv2
```

**Output:**
- Go structs + grpc-gateway code → `gen/go`
- OpenAPI/Swagger spec → `gen/openapiv2`

## Workflow for updating `base.proto`

1. **Edit `base.proto`** — add/change messages, services, or RPCs as needed.
2. **Validate it compiles** (catches import errors, typos, syntax issues):
   ```
   buf build
   ```
3. **(If you changed dependencies)** re-lock them:
   ```
   buf dep update
   ```
   Only needed if you added/changed something in the `deps:` list of `buf.yaml` — not needed for routine proto edits.
4. **Lint (optional but recommended)**:
   ```
   buf lint
   ```
5. **Check for breaking changes (optional, useful before merging)**:
   ```
   buf breaking --against '.git#branch=main'
   ```
6. **Generate code**:
   ```
   buf generate
   ```
   Regenerates everything in `gen/go` and `gen/openapiv2` from the current `.proto` files.
7. **Rebuild/restart your Go service** to pick up the newly generated code.

## Gotchas learned from setup

- `buf.yaml` (v2) requires an explicit `modules:` block pointing at the directory with your `.proto` files — it's not implied.
- `buf build` only validates/compiles — it does **not** generate code. `buf generate` is the one that writes files to disk.
- Remote plugins (`remote:` in `buf.gen.yaml`) run on Buf's servers, so `buf generate` needs network access. If working offline/airgapped, switch to `local:` plugins with the binaries installed via `go install`.
- Plugin versions default to "latest" unless pinned (e.g. `buf.build/protocolbuffers/go:v1.36.6`) — worth pinning before this becomes a shared/production project so generated code doesn't shift under you.

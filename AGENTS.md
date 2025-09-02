# Repository Guidelines

## Project Structure & Module Organization

- `cmd/`: entry points — `subconverter` (server), `subctl` (CLI), `subworker` (background jobs).
- `internal/app/`: core logic — `converter`, `parser`, `generator`.
- `internal/domain/`: entities — `proxy`, `ruleset`, `subscription`.
- `internal/infra/`: integrations — `cache`, `config`, `http`, `storage`.
- `internal/pkg/`: shared utils — `logger`, `errors`, `validator`.
- `configs/`: example and runtime configs. `docs/`: documentation. `test/`: tests. `k8s/`: manifests. `bin/`: build artifacts.

## Build, Test, and Development Commands

- Build: `make build` (binaries in `bin/`). Dev build with race: `make dev`.
- Run locally: `make dev-server`, `make dev-cli`, `make dev-worker`.
- Test: `make test`; coverage HTML: `make test-coverage` (outputs `coverage.html`).
- Lint & format: `make lint`, `make fmt`; tidy modules: `make tidy`.
- Benchmarks: `make benchmark`. Generate code/mocks: `make generate`.
- Docker: `make docker`; compose up/down: `make docker-compose-up` / `make docker-compose-down`.
- Release cross‑compile: `make release`. Install tools: `make install`.

## Coding Style & Naming Conventions

- Formatting: use `go fmt` (tabs, standard Go formatting). Do not hand-format.
- Linting: `golangci-lint` via `make lint`; fix issues before pushing.
- Naming: packages lowercase (no underscores). Exported types/functions in `PascalCase`; unexported in `camelCase`.
- Errors: wrap with context (`fmt.Errorf("...: %w", err)` or project helpers). Prefer returning errors over panics.
- Imports: group stdlib, third‑party, then local (module: `github.com/rogeecn/subconverter-go`).
- Logging: use the project logger (`internal/pkg/logger`).

## Testing Guidelines

- Frameworks: Go `testing` with `testify` assertions. Place tests alongside code in `*_test.go`.
- Naming: `TestXxx` for unit tests; table‑driven where appropriate.
- Running: `make test` (CI runs with `-race -cover`). Coverage report via `make test-coverage`.
- Environment: Redis is optional; when needed, use `REDIS_HOST`/`REDIS_PORT` and provide fakes where possible.

## Commit & Pull Request Guidelines

- Commits: use Conventional Commits (e.g., `feat: ...`, `fix: ...`, `chore: ...`). Example: `feat(parser): add Hysteria2 support`.
- PRs: include a clear description, linked issues (e.g., `Closes #123`), tests for changes, and docs/config updates when relevant.
- Quality gates: ensure `make dev-setup` (tidy, format, lint, test) passes before requesting review. CI also runs security scan (`gosec`).

## Security & Configuration Tips

- Configuration lives in `configs/`; avoid committing secrets. Prefer env vars for credentials.
- Validate inputs (see `internal/pkg/validator`). Be cautious with URL/regex handling and external I/O.

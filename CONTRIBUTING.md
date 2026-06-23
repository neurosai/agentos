# Contributing to AgentOS

Thank you for contributing. AgentOS is early; preserve contract-first design and keep implemented behavior clearly separated from planned behavior.

## Before opening a PR

```bash
make verify
```

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/): `feat`, `fix`, `docs`, `chore`, `refactor`, `test`.

Write commit messages and PR descriptions in **English**.

## Contract sync rule

When changing a public model, update together:

- domain type in `internal/domain/`
- API or schema contract in `api/`
- example in `examples/` (if applicable)
- migration in `migrations/` (if persisted)
- tests
- `docs/implementation-status.md` (if implementation status changes)

## Documentation language

All committed documentation must be **English only**. This includes README files, `docs/`, ADRs, and inline contract descriptions.

## GitHub tracker

Issues, PR titles, descriptions, and comments must be **English only**.

## No in-repo devlog

Do not add development journals (`docs/devlog/`, `DEVLOG.md`, weekly notes). Track progress via GitHub Issues and update `docs/implementation-status.md` only when component status changes.

## Architecture

See `docs/architecture.md` and ADRs in `docs/adr/`. v0.2 runs a single `agentosd` process; legacy `cmd/*d` stubs are not the development target.

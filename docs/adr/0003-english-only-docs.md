# ADR-0003: English-only documentation and no in-repo devlog

## Status

Accepted

## Context

AgentOS is an open-source foundation repository. Contributors and users need consistent, accessible documentation. Progress tracking should not fragment across narrative markdown journals inside the repo.

## Decision

1. **All committed documentation** is English only: `README.md`, `docs/`, ADRs, package READMEs, godoc on exported APIs, OpenAPI descriptions.
2. **GitHub tracker** (issues, PRs, milestones, release notes) is English only.
3. **No in-repo devlog:** do not add `docs/devlog/`, `DEVLOG.md`, weekly status files, or scratch notes under `docs/`.
4. **Allowed progress artifacts:** GitHub Issues/PRs, ADRs for durable decisions, `docs/implementation-status.md` (factual status matrix only).

## Consequences

- PRs with non-English docs are rejected in review.
- `CONTRIBUTING.md` encodes these rules.
- Local scratch files may be gitignored (`*.local.md`, `notes/`) but must not be committed.

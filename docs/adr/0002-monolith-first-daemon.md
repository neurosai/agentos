# ADR-0002: Monolith-first daemon (`agentosd`)

## Status

Accepted

## Context

The Makefile builds seven separate daemon binaries (`taskd`, `policyd`, …). For v0.2 the goal is a governed vertical slice, not multi-process operations (service discovery, inter-service auth, network retries).

## Decision

1. **v0.2:** ship one process `agentosd serve` with internal modules: task, policy, audit, tool, memory.
2. **v0.4+:** optional split mode into separate daemons when deployment semantics are proven.
3. Legacy `cmd/*d` placeholders remain buildable but are not developed in v0.2.

## Consequences

- Simpler local development: one config file, one HTTP port.
- `agentctl` targets `agentosd` base URL.
- Splitting later requires stable internal module boundaries (already enforced via ports/use cases).

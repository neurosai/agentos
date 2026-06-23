# ADR-0001: Contract-first development

## Status

Accepted

## Context

AgentOS defines a control plane for governed agent execution. Multiple services (TaskD, PolicyD, ToolD, MemoryD, AuditD) must interoperate with consistent semantics before distributed deployment.

## Decision

1. Define domain models, OpenAPI/JSON Schema/protobuf contracts, SQL migrations, and OPA policies **before** shipping runnable daemons.
2. Any change to a public model must update domain type, API contract, example, migration (if persisted), and tests together.
3. v0.2 implements existing contract paths only; avoid expanding API surface without ADR.

## Consequences

- Foundation phase (v0.1) produced specifications without runnable services.
- v0.2 implements contracts against a single `agentosd` process.
- Contract tests (`make contracts`) gate regressions.

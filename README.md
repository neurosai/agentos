<div align="center">

```text
    ___                    __  ____  _____
   /   | ____ ____  ____  / /_/ __ \/ ___/
  / /| |/ __ `/ _ \/ __ \/ __/ / / /\__ \
 / ___ / /_/ /  __/ / / / /_/ /_/ /___/ /
/_/  |_\__, /\___/_/ /_/\__/\____//____/
      /____/     GOVERNED AGENT RUNTIME
```

# AgentOS

**A policy-first control plane for running AI agents as governed infrastructure.**

[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache--2.0-blue.svg)](LICENSE)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.1-6BA539?logo=openapiinitiative&logoColor=white)](api/openapi/)
[![OPA](https://img.shields.io/badge/Policy-OPA-7D9199?logo=openpolicyagent&logoColor=white)](policies/)
[![Status](https://img.shields.io/badge/status-v0.1%20foundation-orange)](#project-status)
[![GitHub](https://img.shields.io/badge/GitHub-neurosai%2Fagentos-181717?logo=github)](https://github.com/neurosai/agentos)

</div>

AgentOS is not a new kernel. It is a Linux-based operating environment for AI agents: a system layer for tasks, policy, tools, memory, audit, discovery, and operational context. Linux supplies process isolation, networking, filesystems, cgroups, and namespaces; AgentOS supplies the control plane that makes autonomous work governable.

This repository is the **v0.1 foundation**. It defines the domain model, service boundaries, API contracts, persistence model, and baseline policies. The service binaries are placeholders; AgentOS does not execute agents yet.

## Why AgentOS?

An agent can call tools. An operating environment must also answer:

- Who requested the action, and which agent is acting?
- Is the action allowed for this resource, workspace, and data classification?
- Which credentials may be exchanged, and for what audience?
- What context and memory may the agent read or change?
- Can the action be reconstructed and verified later?
- What services, APIs, resources, and dependencies actually exist?

AgentOS treats these as system responsibilities instead of leaving each agent implementation to solve them independently.

## Architecture

```text
                         humans / CI / agents
                                  |
                           +------+------+
                           |   agentctl  |
                           +------+------+
                                  |
        +-------------------------+-------------------------+
        |                 AGENTOS CONTROL PLANE             |
        |                                                   |
        |  +---------+   authorize   +---------+            |
        |  |  TaskD  |<------------->| PolicyD |<--- OPA    |
        |  +----+----+               +----+----+            |
        |       |                         ^                 |
        |       | run / delegate          | every boundary  |
        |       v                         |                 |
        |  +---------+   invoke      +----+----+            |
        |  | AgentD* |-------------->|  ToolD  |----> MCP   |
        |  +----+----+               +---------+      tools |
        |       |                                           |
        |       +--------------+----------------+           |
        |                      |                |            |
        |                 +----v----+      +----v----+       |
        |                 | MemoryD |      | AuditD |       |
        |                 +----+----+      +----+----+       |
        |                      |                |            |
        |  +------------+  +---v------+         |            |
        |  | DiscoveryD |->| CatalogD |---------+            |
        |  +------+-----+  +----+-----+                      |
        +---------|-------------|----------------------------+
                  |             |
             Git / K8s /     PostgreSQL + pgvector
             CI / OTel

  * AgentD and actual agent runtimes are deferred beyond the foundation.
```

Every privileged boundary is designed to be policy-checked and audit-correlated. Discovery is read-only in v0.1, observed catalog facts remain distinguishable from declared metadata, and memory records carry provenance, ACL, classification, confidence, and retention metadata.

## Components

| Component | Responsibility | Foundation artifact |
| --- | --- | --- |
| **TaskD** | Task lifecycle, messages, approvals, artifacts, event streams | Domain model, ports, use cases, OpenAPI, protobuf, migrations |
| **PolicyD** | Central authorization, obligations, data filters, token-exchange decisions | Domain model, OpenAPI, OPA policies |
| **ToolD** | Governed syscall boundary for MCP and built-in tools | Domain model, ports, OpenAPI, migrations |
| **MemoryD** | Governed memory with provenance, ACLs, retention, and hybrid-search contracts | Domain model, JSON Schema, OpenAPI, pgvector migration |
| **AuditD** | Append-only, hash-linked evidence and proof contracts | Domain model, OpenAPI, migrations |
| **CatalogD** | Typed operational graph of systems, services, APIs, resources, tools, and agents | Domain model, JSON Schema, OpenAPI, migrations |
| **DiscoveryD** | Safe, scoped, read-only collection from Git, Kubernetes, CI/CD, API descriptors, and OTel metadata | Domain model, OPA policy, OpenAPI, migrations |
| **AgentD** | Agent lifecycle and runtime isolation | Deferred; manifest contract only |
| **agentctl** | Operator-facing CLI | Placeholder binary |

## Design principles

- **Policy first.** Tasks, tool calls, memory operations, and discovery requests are evaluated before execution.
- **Identity is end-to-end.** Human, service, parent-agent, and child-agent identities remain explicit across delegation.
- **Tools are syscalls.** MCP and built-in tools are accessed through one governed invocation boundary.
- **Memory is data, not a prompt dump.** Records are typed, attributable, filterable, revisable, forgettable, and retention-aware.
- **Audit is structural.** Events carry task and trace correlation plus hash-chain fields for tamper evidence.
- **Discovery is safe by default.** v0.1 collectors are allowlisted, scoped, and read-only; packet capture, secret reads, network sniffing, and credential guessing are forbidden.
- **Contracts before daemons.** OpenAPI, protobuf, JSON Schema, domain invariants, and persistence shape are stabilized before network implementations.

## Project status

> [!IMPORTANT]
> AgentOS is currently a **foundation/specification repository**, not a runnable agent platform.

### Available now

- Clean Architecture domain, port, and use-case layers
- OpenAPI 3.1 contracts for seven services
- protobuf contracts for task event streaming and catalog ingestion
- JSON Schemas for agent manifests, catalog entities, and memory records
- PostgreSQL migrations for tasks, audit, memory, catalog, discovery, and tools
- pgvector-backed memory storage shape and HNSW index
- OPA/Rego policy bundles with tests
- example task, agent, memory, and catalog documents
- infrastructure Compose profile for PostgreSQL, OPA, and OpenTelemetry Collector
- placeholder CLI and daemon entry points

### Not implemented yet

- HTTP/gRPC servers and persistence adapters
- task orchestration and agent execution
- MCP proxying and credential exchange
- embedding generation and memory retrieval
- discovery collectors and catalog reconciliation
- production authentication and authorization plumbing
- AgentD, Hermes integration, Nix packaging, Vault, and Qdrant

## Quick start

### Prerequisites

- Go 1.23+
- Docker with Compose
- [buf](https://buf.build/docs/installation)
- [goose](https://github.com/pressly/goose)
- [OPA](https://www.openpolicyagent.org/docs/latest/#running-opa)
- [golangci-lint](https://golangci-lint.run/)

### Build and test the foundation

```bash
git clone https://github.com/neurosai/agentos.git
cd agentos

make build
go test ./internal/domain/... ./pkg/...
```

Built placeholders are written to `bin/`:

```bash
./bin/agentctl
./bin/taskd
```

The repository also defines the intended full verification gate:

```bash
make verify
```

This runs Go linting and tests, protobuf linting, OPA policy tests, and contract tests. The migration suite additionally requires a working Docker runtime.

### Start the infrastructure profile

```bash
cd deploy/docker
cp .env.example .env
docker compose up -d
docker compose ps
```

The Compose profile starts infrastructure only. It does not start AgentOS daemons.

### Apply database migrations

```bash
export DATABASE_URL='postgres://agentos:agentos@localhost:5432/agentos?sslmode=disable'
make migrate-up
```

## Contracts and examples

```text
api/openapi/       REST contracts for TaskD, PolicyD, AuditD, ToolD,
                   MemoryD, CatalogD, and DiscoveryD
api/jsonschema/    agent manifest, catalog entity, and memory record schemas
proto/agentos/v1/  task-event and catalog-ingestion streaming contracts
policies/agentos/  baseline Rego policies and tests
examples/          representative agent, task, memory, and catalog documents
```

Example documents:

- [`examples/agents/hermes-dev.yaml`](examples/agents/hermes-dev.yaml) — a future Hermes runtime manifest
- [`examples/tasks/repo-analysis.yaml`](examples/tasks/repo-analysis.yaml) — a task submission
- [`examples/memory/fact.json`](examples/memory/fact.json) — a governed catalog fact
- [`examples/catalog/component-payment-api.yaml`](examples/catalog/component-payment-api.yaml) — a catalog component

## Repository layout

```text
agentos/
|-- api/                 OpenAPI and JSON Schema contracts
|-- cmd/                 placeholder daemon and CLI entry points
|-- deploy/docker/       local infrastructure profile
|-- examples/            example manifests and records
|-- internal/
|   |-- domain/          entities, state machines, and invariants
|   |-- port/            outbound dependency interfaces
|   `-- usecase/         application service boundaries
|-- migrations/          goose migrations for PostgreSQL + pgvector
|-- pkg/                 shared IDs, errors, API generation, and versioning
|-- policies/            OPA/Rego policy bundles and tests
|-- proto/               protobuf streaming contracts
|-- Makefile
`-- README.md
```

## Roadmap

```text
 v0.1 FOUNDATION           v0.2 OPERABLE CORE          v0.3 HARDENED PLATFORM
 +------------------+      +---------------------+     +----------------------+
 | schemas          |      | runnable daemons    |     | AgentD lifecycle     |
 | domain models    | ---> | Postgres adapters   | --> | A2A delegation       |
 | API contracts    |      | OPA enforcement     |     | Nix/NixOS packaging  |
 | policies         |      | MCP tool gateway    |     | Vault + optional     |
 | migrations       |      | memory retrieval    |     | Qdrant               |
 +------------------+      +---------------------+     | signed releases      |
                                                     +----------------------+
```

The immediate engineering goal is a narrow vertical slice:

```text
submit task -> evaluate policy -> execute one agent/tool path
            -> persist result -> append audit evidence -> stream status
```

That slice must work end-to-end before expanding the number of runtimes, collectors, or storage backends.

## Contributing

AgentOS is early. Changes should preserve the contract-first approach and keep implemented behavior clearly separated from planned behavior.

Before opening a pull request:

```bash
make verify
```

When changing a public model, update the corresponding domain type, API/schema contract, example, migration, and tests together.

## License

Licensed under the [Apache License 2.0](LICENSE).

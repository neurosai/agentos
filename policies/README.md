# Policy Bundles

AgentOS v0.1 uses OPA/Rego as the policy decision engine. PolicyD (future) will evaluate these bundles via OPA REST API.

## Decision effects

| Effect | Meaning |
|--------|---------|
| `allow` | Action permitted |
| `deny` | Action rejected |
| `require_approval` | Human approval required before proceeding |
| `redact` | Allow but redact sensitive output fields |
| `filter` | Return filtered query results (Compile API) |
| `exchange_token` | Allow with RFC 8693 token exchange obligation |
| `sandbox_only` | Restrict to sandbox workspace |

## Bundles

| Package | File | Scope |
|---------|------|-------|
| `agentos.tools` | `agentos/tools.rego` | Tool invocation |
| `agentos.memory` | `agentos/memory.rego` | Memory read/write filtering |
| `agentos.discovery` | `agentos/discovery.rego` | Safe collector enforcement |

## Testing

```bash
opa test ./policies/...
```

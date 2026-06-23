# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| 0.2.x   | Yes       |
| 0.1.x   | Foundation only (no runnable control plane) |

## Reporting a vulnerability

Email security reports to the maintainers via GitHub Security Advisories on [neurosai/agentos](https://github.com/neurosai/agentos/security/advisories/new).

Please include:

- Description and impact
- Steps to reproduce
- Affected version or commit

Do not open public issues for undisclosed security vulnerabilities.

## Scope notes

- **DiscoveryD unsafe collectors** (`packet_capture`, `host_scan`, `credential_guess`, `secret_read`, `network_sniff`) are explicitly out of scope and must not be implemented.
- v0.2 uses a **development authentication stub**; production OIDC/Keycloak integration is planned for later releases.

## Response

We aim to acknowledge reports within 5 business days.

# Environment & Access

This page documents the per-environment configuration for Team Health Check: URLs, database connection strings, access groups, and deployment details.

> For local development setup, see [Local Setup Guide](./local-setup.md).

---

## Environment Summary

| Environment | Purpose | Frontend URL | Backend URL | DB Host |
|-------------|---------|-------------|-------------|---------|
| **Local** | Individual dev; full stack on laptop | http://localhost:3000 | http://localhost:8080 | localhost:5432 |
| **Dev** | Shared integration testing; auto-deployed on merge to `main` | TODO: https://teams360-dev.example.com | TODO: https://api-dev.teams360.example.com | TODO: dev-db.internal:5432 |
| **Int (Integration)** | Pre-prod validation; deployed on release candidate tag | TODO: https://teams360-int.example.com | TODO: https://api-int.teams360.example.com | TODO: int-db.internal:5432 |
| **Prod** | Live; deployed on signed release tag (`v*.*.*`) | TODO: https://teams360.example.com | TODO: https://api.teams360.example.com | TODO: prod-db.internal:5432 |

---

## Access Requirements

### Dev Environment

| Access type | How to request | Group / role needed | Approver |
|-------------|---------------|-------------------|---------|
| Read the Dev DB | TODO: IT ticket or Okta group request | TODO: `teams360-dev-db-read` | TODO: tech lead |
| Write to the Dev DB | TODO: IT ticket | TODO: `teams360-dev-db-write` | TODO: tech lead |
| Deploy to Dev | GitHub Actions — automatic on merge to `main` | GitHub write access to repo | TODO |
| VPN / bastion | TODO: IT ticket | TODO: VPN group | TODO: IT |

### Int Environment

| Access type | How to request | Group / role needed | Approver |
|-------------|---------------|-------------------|---------|
| Read the Int DB | TODO: IT ticket | TODO: `teams360-int-db-read` | TODO: senior engineer |
| Deploy to Int | GitHub Actions — automatic on release candidate tag | GitHub write access to repo | TODO |
| View Grafana dashboards | TODO | TODO | TODO |

### Prod Environment

| Access type | How to request | Group / role needed | Approver |
|-------------|---------------|-------------------|---------|
| Read the Prod DB | TODO: IT ticket with business justification | TODO: `teams360-prod-db-read` | TODO: team lead + security |
| Write to the Prod DB | Emergency only — raise incident | TODO | TODO: on-call lead |
| Deploy to Prod | GitHub Actions — automatic on signed `v*.*.*` tag | Maintainer role on repo | TODO |

---

## Environment Variables Per Environment

### Local

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/teams360?sslmode=disable
API_PORT=8080
GIN_MODE=debug
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Dev

```env
DATABASE_URL=postgres://TODO_USER:TODO_PASSWORD@TODO_HOST:5432/teams360?sslmode=require
API_PORT=8080
GIN_MODE=release
NEXT_PUBLIC_API_URL=https://api-dev.teams360.example.com

# SSO (if configured in Dev)
OAUTH_CLIENT_ID=TODO
OAUTH_TOKEN_URL=https://TODO_PROVIDER/oauth/token
OAUTH_REDIRECT_URI=https://teams360-dev.example.com/auth/callback
```

### Int / Prod

Same as Dev with environment-specific values. Never share Prod credentials over Slack or email. Use the secrets manager (TODO: name of secrets manager — Vault / AWS Secrets Manager / etc.).

---

## Database Connection Strings

Use these only after receiving access from the approver above.

```bash
# Local
psql "postgres://postgres:postgres@localhost:5432/teams360?sslmode=disable"

# Dev (requires VPN)
psql "postgres://TODO_USER:TODO_PASSWORD@TODO_DEV_HOST:5432/teams360?sslmode=require"

# Int (requires VPN + approval)
psql "postgres://TODO_USER:TODO_PASSWORD@TODO_INT_HOST:5432/teams360?sslmode=verify-full"

# Prod (emergency read only)
psql "postgres://TODO_USER:TODO_PASSWORD@TODO_PROD_HOST:5432/teams360?sslmode=verify-full"
```

---

## SSO Configuration Per Environment

Team Health Check supports OIDC/OAuth 2.0 (PKCE flow). Provider must be configured with:

| Setting | Value |
|---------|-------|
| Application type | Single Page App (SPA) / Public client |
| Client secret | Not required (PKCE) |
| Allowed redirect URI (Local) | `http://localhost:3000/auth/callback` |
| Allowed redirect URI (Dev) | `https://teams360-dev.example.com/auth/callback` |
| Allowed redirect URI (Int) | `https://teams360-int.example.com/auth/callback` |
| Allowed redirect URI (Prod) | `https://teams360.example.com/auth/callback` |
| Required scopes | `openid email profile` |
| Token claim needed | `email` in ID token |

TODO: Confirm Okta / Keycloak / other provider and application names per environment.

---

## Escalation Contacts

| Issue | Contact | Channel |
|-------|---------|---------|
| Dev DB access | TODO: team lead name | `#team360-eng` |
| Prod incident | TODO: on-call rotation | `#team360-incidents` |
| SSO / auth issue | TODO: security team | `#security-help` |
| GitHub Actions failure | TODO: DevOps / platform team | `#devops-help` |
| General questions | TODO: team Slack handle | `#team360-eng` |

---

## Questions for Product / Tech Leads

- TODO: Confirm exact Dev / Int / Prod URLs.
- TODO: Confirm the secrets management solution (Vault, AWS Secrets Manager, etc.).
- TODO: Confirm SSO provider and Okta/Keycloak application names for each environment.
- TODO: Confirm access approval process and SLA for Dev DB access requests.
- TODO: Is there an on-call rotation for Prod incidents? Who owns it?

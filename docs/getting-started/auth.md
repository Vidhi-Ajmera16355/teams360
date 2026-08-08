# Authentication & Authorization

This page covers how Team Health Check authenticates users and enforces role-based access control (RBAC).

---

## Authentication Methods

Team Health Check supports two authentication methods simultaneously. Both coexist — enabling one does not disable the other.

| Method | Description | When to use |
|--------|-------------|------------|
| **Username / password** | Credentials stored in `users.password_hash` (bcrypt) | Default; always available |
| **SSO / OIDC (PKCE)** | Authorization Code flow with PKCE; no client secret required | Enabled by setting OAuth env vars |

---

## Username / Password Flow

```
1. User submits credentials → POST /api/v1/auth/login
2. auth_handler.go calls UserRepository.FindByUsername()
3. bcrypt.CompareHashAndPassword(storedHash, submittedPassword)
4. On match: return user object + set-cookie header
5. Frontend (js-cookie) stores userId in cookie (expires 1 day)
6. middleware.ts reads cookie on every page load
7. Missing or invalid cookie → redirect to /login
```

**Password hashing:** bcrypt with cost factor of 12 (Go `golang.org/x/crypto/bcrypt`). Passwords are never stored in plain text.

**Demo passwords:** All demo accounts use the string `"demo"` as the password (hashed). Admin uses `"admin"`. These are seeded via migration `000007`. Never use demo credentials in production.

---

## SSO / OIDC Flow (Authorization Code + PKCE)

Team Health Check supports any OIDC-compliant provider: Keycloak, Okta, Auth0, Google, Azure AD, etc.

```
1. User clicks "Sign in with SSO" on /login
2. Frontend generates PKCE code_verifier + code_challenge
3. Frontend redirects to provider's /authorize endpoint:
      ?client_id=...
      &redirect_uri=.../auth/callback
      &response_type=code
      &scope=openid email profile
      &code_challenge=...
      &code_challenge_method=S256
4. User authenticates at provider
5. Provider redirects to /auth/callback?code=...
6. Frontend calls POST /api/v1/auth/sso/callback
      { code, code_verifier, redirect_uri }
7. Backend calls provider's token endpoint to exchange code → tokens
8. Backend extracts `email` from ID token claims
9. Backend looks up user by email in users table
10. If found: same cookie-set flow as username/password
11. If not found: return 401 "SSO account not provisioned"
    (Team Health Check does not auto-create users from SSO logins)
```

### Configure SSO

**Frontend** (`frontend/.env.local`):

```env
NEXT_PUBLIC_OAUTH_CLIENT_ID=your-client-id
NEXT_PUBLIC_OAUTH_AUTHORIZE_URL=https://your-provider.com/oauth/authorize
NEXT_PUBLIC_OAUTH_REDIRECT_URI=http://localhost:3000/auth/callback
NEXT_PUBLIC_OAUTH_SCOPES=openid email profile
```

**Backend** (`backend/.env`):

```env
OAUTH_CLIENT_ID=your-client-id
OAUTH_TOKEN_URL=https://your-provider.com/oauth/token
OAUTH_REDIRECT_URI=http://localhost:3000/auth/callback
```

If the frontend OAuth env vars are not set, the "Sign in with SSO" button is hidden.

### Provider setup quick reference

| Setting | Value |
|---------|-------|
| Application type | Single Page App (SPA) / Public client |
| Client secret | Not required |
| Allowed redirect URI | `http://localhost:3000/auth/callback` (or your domain) |
| Required scopes | `openid email profile` |
| Token claim needed | `email` in ID token |

---

## Role-Based Access Control (RBAC)

### Hierarchy Levels

Permissions are tied to `hierarchy_levels`, not to individual users. Each level has boolean flags:

| Permission flag | VP (L1) | Director (L2) | Manager (L3) | Team Lead (L4) | Team Member (L5) | Admin (L0) |
|----------------|---------|--------------|-------------|---------------|-----------------|-----------|
| `can_view_all_teams` | true | true | false | false | false | true |
| `can_edit_teams` | false | false | false | false | false | true |
| `can_manage_users` | false | false | false | false | false | true |
| `can_take_survey` | false | false | false | true | true | false |
| `can_view_analytics` | true | true | true | true | false | true |
| `can_configure_system` | false | false | false | false | false | true |
| `can_view_reports` | true | true | true | false | false | true |
| `can_export_data` | true | true | false | false | false | true |

These defaults can be modified via the Admin panel (Hierarchy Levels tab) without a code change.

### Route Protection

`frontend/middleware.ts` enforces routes based on the cookie:

```typescript
// Simplified logic
if (!userId) return redirect('/login');

const user = await getUser(userId);

if (path.startsWith('/admin') && !user.permissions.canConfigureSystem) {
  return redirect('/home');
}

if (path.startsWith('/dashboard') && !user.permissions.canTakeSurvey) {
  return redirect('/manager');
}
// ... etc.
```

### Team Access Control

A user can access a team's data if **any** of the following is true:

1. `can_view_all_teams` is true for their hierarchy level (VP, Director, Admin)
2. The user is a member of the team (`team_members` table)
3. The user appears in the team's `team_supervisors` chain

Logic implemented in `frontend/lib/org-config.ts`:

```typescript
export function canUserAccessTeam(user: User, team: Team, orgConfig: OrgConfig): boolean {
  const permissions = getUserPermissions(user, orgConfig);
  if (permissions.canViewAllTeams) return true;
  if (team.memberIds.includes(user.id)) return true;
  return team.supervisorChain.some(s => s.userId === user.id);
}
```

---

## Password Reset

Team Health Check implements a short-lived (1-hour), single-use token flow:

```
1. User submits email on /forgot-password
2. Backend creates a password_reset_tokens row (token_hash stored, not plaintext)
3. Email sent with reset link (token in URL)
4. User clicks link → token verified → password updated → token marked used
```

Token table: `password_reset_tokens` — see [DB Schema](../architecture/db-schema.md#password_reset_tokens).

**Note:** The reset endpoint is not yet rate-limited.

---

## Cookie Specification

| Attribute | Value |
|-----------|-------|
| Name | `userId` (and optionally role metadata) |
| Storage | js-cookie (client-side) |
| Expiry | 1 day |
| SameSite | Lax (default for js-cookie) |
| Secure | TODO: enforce `Secure` flag in production |
| HttpOnly | Not currently set (client JS reads the cookie) |

**Roadmap:** Migrate from cookie-based userId to JWT tokens with refresh token rotation. See `CLAUDE.md` migration roadmap.

---

## Future Improvements

- [ ] JWT tokens with short expiry + refresh token rotation
- [ ] Role-based API middleware in Go (currently enforced only on frontend)
- [ ] Rate limiting on auth endpoints (see Known Issues #8)
- [ ] `HttpOnly` and `Secure` cookie flags
- [ ] HTTPS enforcement middleware in Gin for production

---

## Questions for Product / Tech Leads

- TODO: Confirm which OIDC provider is used per environment (Okta / Keycloak / other).
- TODO: Should Admin-level users be allowed to take surveys? Currently their level has `can_take_survey = false`.
- TODO: Is there a compliance requirement for session timeout (e.g., 8 hours idle)?
- TODO: Confirm whether `HttpOnly` cookie flag will break any existing integrations before enabling.

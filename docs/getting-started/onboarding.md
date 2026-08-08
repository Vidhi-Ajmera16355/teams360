# Onboarding Checklist for New Developers

Welcome to Team Health Check. This checklist walks you through Day 1 to Day 5. Complete each item in order. The goal: by end of Day 5 you understand the system, have a working local stack, and have made your first commit.

Assign yourself as reviewer to a buddy engineer who will sign off the checklist at the end.

---

## Before Day 1 — Access Provisioning

Request these before your first day (or on Day 1 morning). Estimated approval time: 2–24 hours.

- [ ] GitHub access: request read + write access to `github.com/guidewire-oss/teams360`
- [ ] Slack: join `#team360-eng`, `#team360-bugs`, `#team360-frontend`, `#team360-backend`
- [ ] TODO: Request access to the Dev environment database (see [Environment & Access](./environments.md))
- [ ] TODO: Request VPN / bastion access if required for Dev environment
- [ ] TODO: Request Jira project access — board name: TODO
- [ ] TODO: Request GitHub Actions view access if org-level permissions are restricted
- [ ] TODO: If SSO is configured, confirm your corporate email is provisioned in the Dev environment's Team Health Check user table

---

## Day 1 — Setup & Orientation

### Morning: Local environment

- [ ] Read [Overview](../overview.md) — understand what Team Health Check does and who uses it
- [ ] Follow [Local Setup Guide](./local-setup.md) completely — do not skip steps
- [ ] Verify setup validation: log in as `demo/demo`, submit a survey, then log in as `manager1/demo` and confirm health data appears
- [ ] Run backend tests: `cd backend && go test ./...` — all should pass
- [ ] Run E2E tests: `make test-e2e` — all should pass (takes 2–5 minutes)
- [ ] Bookmark internal URLs for Dev environment (see [environments.md](./environments.md))

### Afternoon: Codebase orientation

- [ ] Read [Architecture & Key Flows](../architecture/overview.md) — understand frontend, backend, and DB layers
- [ ] Open `backend/cmd/api/main.go` — trace how routes are registered
- [ ] Open `frontend/middleware.ts` — trace how role-based routing works
- [ ] Open `frontend/lib/types.ts` — familiarize yourself with core data models
- [ ] Ask your buddy to explain any architecture decisions that are unclear

---

## Day 2 — Database & Domain Model

- [ ] Read [Database Provisioning & Configuration](../architecture/db-schema.md) — know all 11 tables
- [ ] Read [Domain Model](../architecture/domain-model.md) — understand DDD layers
- [ ] Connect to local DB and run these queries to understand data shape:
  ```bash
  psql "postgres://postgres:postgres@localhost:5432/teams360?sslmode=disable"
  \dt                              -- list all tables
  SELECT * FROM hierarchy_levels;  -- see org levels
  SELECT * FROM health_dimensions; -- see 11 dimensions
  SELECT * FROM users LIMIT 5;     -- see demo users
  SELECT * FROM teams;             -- see demo teams
  SELECT * FROM team_supervisors WHERE team_id = 'platform-squad';
  ```
- [ ] Trace a health check session end-to-end in the DB:
  1. Submit a survey as `demo/demo`
  2. Find the session: `SELECT * FROM health_check_sessions ORDER BY created_at DESC LIMIT 1`
  3. Find its responses: `SELECT * FROM health_check_responses WHERE session_id = '<id from above>'`
- [ ] Read [Authentication & Authorization](./auth.md) — understand cookie flow and RBAC

---

## Day 3 — Backend Deep Dive

- [ ] Trace the survey submission flow in code:
  - `frontend/app/survey/page.tsx` → POST → `frontend/app/api/v1/health-checks/route.ts`
  - `backend/interfaces/api/v1/health_check_handler.go` → `SubmitHealthCheck()`
  - `backend/infrastructure/persistence/postgres/health_check_repository.go` → `Save()`
- [ ] Read `backend/interfaces/api/v1/auth_handler.go` — understand login and SSO callback handlers
- [ ] Read one complete repository implementation:
  ```
  backend/infrastructure/persistence/postgres/user_repository.go
  ```
- [ ] Understand Ginkgo test structure by reading:
  ```
  tests/acceptance/e2e_survey_submission_test.go
  ```
- [ ] Run a single E2E test suite and observe output:
  ```bash
  cd tests
  export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/teams360_test?sslmode=disable"
  ginkgo -v -focus="E2E: Survey Submission" acceptance/
  ```

---

## Day 4 — Frontend Deep Dive

- [ ] Trace the manager dashboard load in code:
  - `frontend/app/manager/page.tsx` → fetches `/api/v1/managers/{id}/teams/health`
  - `frontend/lib/org-config.ts` → `canUserAccessTeam()` filters teams client-side
  - `backend/interfaces/api/v1/manager_handler.go` → `GetManagerTeamsHealth()`
- [ ] Read `frontend/lib/assessment-period.ts` — understand automatic period detection
- [ ] Read `frontend/lib/org-config.ts` — understand supervisor-chain access control
- [ ] Open the Admin panel as `admin/admin` and explore:
  - Hierarchy Levels tab — try reordering levels
  - Teams tab — create a test team, assign members
  - Dimensions tab — toggle a dimension inactive and see it disappear from the survey
- [ ] Read the Makefile documentation: [Makefile Reference](../development/makefile.md)
- [ ] Review [CI/CD & Deployment](../operations/deployment.md) — understand the GitHub Actions pipeline

---

## Day 5 — First Contribution

- [ ] Pick a "good first issue" from the Jira board (ask your buddy)
- [ ] Create a feature branch:
  ```bash
  git checkout -b your-name/issue-description
  ```
- [ ] Write a failing test first (Ginkgo for backend, or Playwright for E2E):
  ```bash
  # For backend unit/integration:
  cd backend && ginkgo -v -focus="your test description" ./...

  # For E2E:
  cd tests && ginkgo -v -focus="E2E: Your Feature" acceptance/
  ```
- [ ] Implement the fix until tests pass
- [ ] Run the full test suite:
  ```bash
  make test          # backend unit tests
  make test-e2e      # E2E tests
  ```
- [ ] Open a Pull Request with:
  - Link to Jira ticket
  - Description of what changed and why
  - Screenshot or curl output demonstrating the fix
- [ ] Request review from your buddy and one other engineer

---

## Onboarding Sign-Off

Your buddy engineer signs off when:

- [ ] Local stack runs without errors
- [ ] Developer can explain the three critical flows (auth, survey submission, manager dashboard) without the docs
- [ ] Developer has merged at least one PR (even a small fix)
- [ ] Developer knows how to run E2E tests locally
- [ ] Developer knows where to find known issues and how to add a new one

**Buddy sign-off**: ___________________________________ Date: ___________

---

## Useful References

| Resource | URL / Path |
|----------|------------|
| Repo | `github.com/guidewire-oss/teams360` |
| Local frontend | http://localhost:3000 |
| Local backend | http://localhost:8080 |
| Dev environment | TODO — see [environments.md](./environments.md) |
| Jira board | TODO |
| Slack | `#team360-eng` |
| Grafana (local) | http://localhost:3001 (after `make run-with-otel`) |
| API docs | [api/endpoints.md](../api/endpoints.md) |
| Makefile reference | [development/makefile.md](../development/makefile.md) |

---

## Questions for Product / Tech Leads

- TODO: Confirm the "good first issue" label in Jira so new starters know where to start.
- TODO: Is there a mandatory security training or compliance acknowledgement before Dev DB access is granted?
- TODO: Confirm the buddy / mentor assignment process for new team members.
- TODO: Should new developers have write access to the main branch immediately, or only after 30 days?

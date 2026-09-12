# Assignment Submission — ATHENA AML

Repository: <https://github.com/bmaulanaa61-eng/athena-aml>

Live compliance dashboard: <https://datalabs-systems.web.app/app/omnishield>

## Final result

ATHENA AML is a production-minded AML and fraud monitoring application. It accepts transaction ingestion through a Go REST API, stores work in PostgreSQL, and processes AML risk asynchronously through a Python worker. Risk decisions combine deterministic rules with optional OpenAI-compatible enrichment while remaining fully runnable without external AI credentials. A live React compliance dashboard is provided to demonstrate the analyst-facing monitoring, alert triage, investigation, rules, and webhook workflows.

## Requirement checklist

| Requirement | Implementation |
|---|---|
| Simple application, self-defined scope | B2B AML/fraud transaction monitoring service with ingestion API, risk worker, alerts, rules, investigation endpoints, and analyst dashboard demo. |
| Production-minded and maintainable by other engineers | Layered Go API, isolated Python worker, PostgreSQL schema, API auth, idempotency, Docker images, architecture notes, OpenAPI spec, CI, live UI demo, and clear local/deployment docs. |
| AI-assisted development | Built with an AI-assisted workflow using Cline/Cursor-like assistance for scaffolding, refactoring, test generation, Dockerization, and documentation. Architecture and correctness were validated by automated tests and manual review. |
| Git and GitHub workflow | Standalone GitHub repository on `main`, Conventional Commit style, CI on push/PR, recommended feature-branch + pull-request workflow documented in `README.md`. |
| Runnable using Docker | `docker compose up --build` starts PostgreSQL, Go API, and Python worker. |
| Automated testing | Go unit tests via `go test ./...`; Python tests via `pytest`; GitHub Actions runs both and builds Docker images. |
| Frontend/user-facing demo | Live dashboard at <https://datalabs-systems.web.app/app/omnishield> shows KPI monitoring, alert trend, risk distribution, priority queue, and 360° investigation UX. |
| Documentation for handoff | `README.md`, `ARCHITECTURE.md`, `SUBMISSION.md`, `backend/api/openapi.yaml`, `.env.example` files, and inline package structure. |
| No VPS/server provided | Deployment approach recommends managed PostgreSQL plus serverless/container platforms such as Cloud Run, Render, Fly.io, Railway, Neon, or Supabase. |
| Timebox-friendly result | Core repo scope stays backend-focused: API + worker + database + Docker + tests + CI + docs, complemented by a live frontend demo link. |

## How to run locally

Prerequisite: Docker Desktop.

```bash
docker compose up --build
```

Health check:

```bash
curl http://localhost:8080/v1/health
```

Seeded demo API key:

```text
X-API-KEY: os_test_demo_super_secret_key_please_change
```

Example risky transaction:

```bash
curl -X POST http://localhost:8080/v1/transactions/ \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: os_test_demo_super_secret_key_please_change" \
  -H "Idempotency-Key: demo-risky-001" \
  -d '{
    "external_transaction_id": "ext-risky-001",
    "account_id": "acct-1001",
    "counterparty_id": "new-counterparty-9999",
    "amount": 75000000,
    "currency": "IDR",
    "transaction_type": "P2P_TRANSFER",
    "occurred_at": "2026-09-12T02:30:00Z",
    "metadata": {"channel": "mobile", "device_id": "android-emulator"}
  }'
```

Then check generated alerts:

```bash
curl http://localhost:8080/v1/alerts/ \
  -H "X-API-KEY: os_test_demo_super_secret_key_please_change"
```

## How to run tests

```bash
make test
```

Manual alternative:

```bash
cd backend && go test ./...
cd ../worker && python -m pytest -q
```

## CI verification

GitHub Actions workflow: `.github/workflows/ci.yml`

CI jobs:

1. Backend: install Go, download modules, run `go test ./...`, build API Docker image.
2. Worker: install Python dependencies, run `pytest`, build worker Docker image.

Latest verified CI run:

- Workflow: `CI`
- Branch: `main`
- Event: `push`
- Conclusion: `success`
- URL: <https://github.com/bmaulanaa61-eng/athena-aml/actions/runs/34692024876>

## GitHub repository setup

Configured repository metadata:

- Description: `Production-ready AML platform with Go API, Python risk worker, PostgreSQL queue, Docker Compose, tests, and CI.`
- Topics: `aml`, `backend`, `docker`, `fraud-detection`, `github-actions`, `go`, `microservices`, `postgresql`, `python`, `risk-engine`

Recommended branch workflow for future engineers:

1. Create issue or task.
2. Create branch: `feat/<short-name>`, `fix/<short-name>`, `docs/<short-name>`, or `test/<short-name>`.
3. Commit using Conventional Commits.
4. Open PR into `main`.
5. Merge only after CI passes and review is complete.
6. Tag releases using SemVer, for example `v0.1.0`.

## Deployment approach without VPS

Because no VPS/server is provided, the safest production path is managed infrastructure:

- Database: Neon, Supabase, Railway PostgreSQL, or another managed PostgreSQL provider.
- API: Cloud Run, Render Web Service, Fly.io, or Railway Service using the backend Dockerfile.
- Worker: Cloud Run service/job, Render Background Worker, Fly.io process group, or Railway Worker using the worker Dockerfile.
- Secrets: provider secret manager for database credentials and optional `AI_API_KEY`.
- CI/CD: extend GitHub Actions to build/push images and deploy after tests pass.

## Notes for the next engineer

- The AI enrichment is optional by design. If `AI_API_KEY` is absent, the worker uses deterministic fallback behavior so local Docker and CI remain stable.
- PostgreSQL currently uses compose init scripts. A production rollout should introduce versioned migrations such as Goose, Flyway, or Atlas.
- Future hardening candidates: request rate limiting, structured logging, OpenTelemetry, integration tests with Testcontainers, coverage reporting, and branch protection/rulesets.

# ATHENA AML — Transaction Intelligence & Compliance Platform

ATHENA AML adalah aplikasi B2B SaaS sederhana namun production-minded untuk menerima transaksi finansial, mengevaluasi risiko AML/fraud secara asynchronous, dan membuat alert investigasi untuk compliance analyst.

Project ini dibuat sebagai hasil **AI-assisted development** menggunakan Cline/Cursor-like workflow: AI digunakan untuk scaffolding, refactor, test generation, Dockerization, dan dokumentasi; keputusan arsitektur tetap divalidasi lewat test otomatis dan review manual.

## Fitur utama

- **Go Ingestion API**: REST API high-throughput menggunakan `net/http`, `chi`, dan `pgx`.
- **PostgreSQL multi-tenant schema**: UUID PK, JSONB, tenant isolation, audit-friendly tables.
- **API key auth**: prefix lookup + bcrypt verification; secret tidak disimpan plaintext.
- **Idempotency engine**: `Idempotency-Key` + SHA-256 payload hash untuk mencegah double-billing.
- **Python AML Worker**: PostgreSQL queue dengan `FOR UPDATE SKIP LOCKED`.
- **Hybrid decisioning**: deterministic rule engine + optional OpenAI-compatible LLM enrichment.
- **Offline-safe AI fallback**: tanpa API key, worker tetap berjalan dengan deterministic fallback.
- **Automated testing**: Go unit tests + Python pytest.
- **Docker-ready**: `docker compose up --build` menjalankan PostgreSQL, API, dan worker.
- **GitHub Actions CI**: test backend, test worker, dan Docker build pada push/PR.

## Tech stack

| Layer | Teknologi |
|---|---|
| API | Go 1.25, chi, pgx |
| Worker | Python 3.11, Pydantic, psycopg, OpenAI-compatible API |
| Database | PostgreSQL 16 |
| Container | Docker, Docker Compose |
| CI | GitHub Actions |

## Quickstart lokal

> Prasyarat: Docker Desktop aktif.

```bash
docker compose up --build
```

API akan tersedia di:

```text
http://localhost:8080/v1/health
```

Demo credentials dari seed:

```text
X-API-KEY: os_test_demo_super_secret_key_please_change
Dashboard email: admin@athena.local
Dashboard password: Admin123!
```

### Health check

```bash
curl http://localhost:8080/v1/health
```

### Ingest transaksi normal

```bash
curl -X POST http://localhost:8080/v1/transactions/ \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: os_test_demo_super_secret_key_please_change" \
  -H "Idempotency-Key: demo-normal-001" \
  -d '{
    "external_transaction_id": "ext-normal-001",
    "account_id": "acct-1001",
    "counterparty_id": "merchant-2001",
    "amount": 125000,
    "currency": "IDR",
    "transaction_type": "PURCHASE",
    "metadata": {"channel": "mobile", "city": "Jakarta"}
  }'
```

### Ingest transaksi berisiko

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

Tunggu beberapa detik agar worker memproses job, lalu cek alert:

```bash
curl http://localhost:8080/v1/alerts/ \
  -H "X-API-KEY: os_test_demo_super_secret_key_please_change"
```

## Menjalankan automated tests

### Semua test

```bash
make test
```

Jika `make` tidak tersedia di Windows, jalankan manual:

```bash
cd backend && go test ./...
cd ../worker && python -m pytest -q
```

## Environment variables

Backend: lihat `backend/.env.example`.
Worker: lihat `worker/.env.example`.

`AI_API_KEY` bersifat opsional. Jika kosong, worker menggunakan offline deterministic fallback agar demo dan test tetap stabil tanpa vendor dependency.

## API documentation

OpenAPI spec tersedia di `backend/api/openapi.yaml`.

Endpoint utama:

- `GET /v1/health`
- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/transactions/`
- `GET /v1/transactions/`
- `GET /v1/transactions/{id}`
- `GET /v1/alerts/`
- `GET /v1/alerts/{id}`
- `POST /v1/alerts/{id}/notes`
- `POST /v1/alerts/{id}/resolve`
- `POST /v1/rules/`
- `GET /v1/rules/`
- `POST /v1/webhooks/`

## Workflow Git & GitHub yang disarankan

- Branch utama: `main`, protected branch di GitHub.
- Development via feature branch: `feat/...`, `fix/...`, `docs/...`, `test/...`.
- Commit menggunakan Conventional Commits:
  - `feat: add AML worker persistence`
  - `test: cover transaction ingestion validation`
  - `ci: run backend and worker checks`
- Pull request wajib hijau di CI sebelum merge.
- Release dapat memakai tag semantic versioning: `v0.1.0`, `v0.2.0`.

## Deployment tanpa VPS

Karena tidak disediakan VPS/server, deployment paling praktis:

1. **Database managed**: Neon, Supabase, atau Railway PostgreSQL.
2. **API container**: Google Cloud Run, Fly.io, Render Web Service, atau Railway Service.
3. **Worker container**: Cloud Run Jobs/Services, Render Worker, Fly.io process group, atau Railway Worker.
4. **Secrets**: simpan `DB_*` dan `AI_API_KEY` di secret manager provider, bukan di Git.
5. **CI/CD**: GitHub Actions build image lalu deploy ke provider container.

## Catatan production hardening berikutnya

- Tambah migration tool versi eksplisit seperti Goose/Flyway.
- Tambah structured logging dan OpenTelemetry tracing.
- Tambah rate limiting per API key.
- Tambah integration test dengan Testcontainers/PostgreSQL.
- Tambah webhook delivery worker real retry/backoff.
- Tambah dashboard frontend untuk analyst workflow.

## License

MIT.

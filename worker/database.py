import os
import json
from typing import Any, Dict

import psycopg
from psycopg.rows import dict_row


class Database:
    def __init__(self):
        self.conninfo = os.getenv("DATABASE_URL") or (
            f"postgresql://{os.getenv('DB_USER')}:{os.getenv('DB_PASSWORD')}@"
            f"{os.getenv('DB_HOST')}:{os.getenv('DB_PORT')}/{os.getenv('DB_NAME')}"
            f"?sslmode={os.getenv('DB_SSLMODE', 'disable')}"
        )
        self.conn = None

    async def connect(self):
        self.conn = await psycopg.AsyncConnection.connect(self.conninfo, row_factory=dict_row)

    async def close(self):
        if self.conn is not None:
            await self.conn.close()

    async def claim_job(self):
        """Atomically claim one AML job using PostgreSQL row locking."""
        async with self.conn.cursor() as cur:
            await self.conn.execute("BEGIN;")
            try:
                query = """
                    SELECT id::text, tenant_id::text, payload, attempts
                    FROM jobs
                    WHERE status IN ('PENDING', 'RETRYING')
                      AND job_type = 'AML_ASSESSMENT'
                      AND available_at <= NOW()
                      AND attempts < max_attempts
                    ORDER BY created_at ASC
                    FOR UPDATE SKIP LOCKED
                    LIMIT 1;
                """
                await cur.execute(query)
                job = await cur.fetchone()
                if not job:
                    await self.conn.execute("COMMIT;")
                    return None

                await cur.execute(
                    """
                    UPDATE jobs
                    SET status = 'PROCESSING', attempts = attempts + 1, locked_at = NOW(), updated_at = NOW()
                    WHERE id = %s::uuid;
                    """,
                    (job["id"],),
                )
                await self.conn.execute("COMMIT;")
                return job
            except Exception:
                await self.conn.execute("ROLLBACK;")
                raise

    async def fetch_transaction(self, tenant_id: str, transaction_id: str) -> Dict[str, Any]:
        async with self.conn.cursor() as cur:
            await cur.execute(
                """
                SELECT id::text, tenant_id::text, external_transaction_id, account_id, counterparty_id,
                       amount::float AS amount, currency, transaction_type, occurred_at, metadata, status
                FROM transactions
                WHERE tenant_id = %s::uuid AND id = %s::uuid
                LIMIT 1;
                """,
                (tenant_id, transaction_id),
            )
            row = await cur.fetchone()
            if row is None:
                raise ValueError(f"transaction {transaction_id} not found for tenant {tenant_id}")
            return dict(row)

    async def save_assessment(self, job_id: str, record):
        deterministic = record.deterministic
        ai = record.ai_enrichment
        ai_verdict = "CLEAR"
        ai_confidence = 0.0
        ai_explanation = None
        risk_factors = []

        if ai is not None:
            verdict_map = {"APPROVE": "CLEAR", "CLEAR": "CLEAR", "REVIEW": "REVIEW", "ESCALATE": "ESCALATE"}
            ai_verdict = verdict_map.get(ai.assessment.upper(), "REVIEW")
            ai_confidence = ai.confidence_score
            ai_explanation = ai.explanation
            risk_factors = ai.risk_factors

        if deterministic.risk_level == "HIGH" and ai_verdict == "CLEAR":
            ai_verdict = "REVIEW"
        elif deterministic.risk_level == "MEDIUM" and ai_verdict == "CLEAR":
            ai_verdict = "REVIEW"

        rules_triggered = [rule.model_dump() for rule in deterministic.triggered_rules]
        priority = "P1" if deterministic.risk_score >= 70 else "P2"

        async with self.conn.cursor() as cur:
            await self.conn.execute("BEGIN;")
            try:
                await cur.execute(
                    """
                    INSERT INTO risk_assessments (
                        tenant_id, transaction_id, risk_score, risk_level, rules_triggered,
                        risk_factors, ai_verdict, ai_confidence, ai_explanation, ai_model, prompt_version
                    ) VALUES (%s::uuid, %s::uuid, %s, %s, %s::jsonb, %s::jsonb, %s, %s, %s, %s, %s)
                    RETURNING id::text;
                    """,
                    (
                        record.tenant_id,
                        record.transaction_id,
                        deterministic.risk_score,
                        deterministic.risk_level,
                        json.dumps(rules_triggered),
                        json.dumps(risk_factors),
                        ai_verdict,
                        ai_confidence,
                        ai_explanation,
                        record.model_used,
                        "aml-worker-v1",
                    ),
                )
                assessment = await cur.fetchone()
                assessment_id = assessment["id"]

                new_status = "ASSESSED"
                if deterministic.risk_score >= 30:
                    await cur.execute(
                        """
                        INSERT INTO alerts (tenant_id, transaction_id, risk_assessment_id, risk_score, risk_level, priority, status)
                        VALUES (%s::uuid, %s::uuid, %s::uuid, %s, %s, %s, 'OPEN');
                        """,
                        (
                            record.tenant_id,
                            record.transaction_id,
                            assessment_id,
                            deterministic.risk_score,
                            deterministic.risk_level,
                            priority,
                        ),
                    )
                    new_status = "REVIEW"

                await cur.execute(
                    """
                    UPDATE transactions
                    SET status = %s, updated_at = NOW()
                    WHERE id = %s::uuid AND tenant_id = %s::uuid;
                    """,
                    (new_status, record.transaction_id, record.tenant_id),
                )
                await cur.execute(
                    """
                    UPDATE jobs
                    SET status = 'COMPLETED', completed_at = NOW(), updated_at = NOW(), last_error = NULL
                    WHERE id = %s::uuid;
                    """,
                    (job_id,),
                )
                await self.conn.execute("COMMIT;")
                return assessment_id
            except Exception:
                await self.conn.execute("ROLLBACK;")
                raise

    async def mark_job_failed(self, job_id: str, error_message: str):
        async with self.conn.cursor() as cur:
            await self.conn.execute("BEGIN;")
            try:
                await cur.execute(
                    """
                    UPDATE jobs
                    SET status = CASE WHEN attempts >= max_attempts THEN 'FAILED' ELSE 'RETRYING' END,
                        available_at = NOW() + INTERVAL '30 seconds',
                        last_error = %s,
                        updated_at = NOW()
                    WHERE id = %s::uuid;
                    """,
                    (error_message[:1000], job_id),
                )
                await self.conn.execute("COMMIT;")
            except Exception:
                await self.conn.execute("ROLLBACK;")
                raise

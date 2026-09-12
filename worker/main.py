import asyncio
import json
import logging
from dotenv import load_dotenv

load_dotenv()

from database import Database
from rule_engine import RuleEngine
from context_builder import ContextBuilder
from ai_client import AIClient
from orchestrator import AIOrchestrator

logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s")
logger = logging.getLogger("AMLWorker")


async def process_job(db: Database, job: dict, rule_engine: RuleEngine, context_builder: ContextBuilder, orchestrator_agent: AIOrchestrator):
    job_id = job["id"]
    tenant_id = job["tenant_id"]
    payload = job["payload"] if isinstance(job["payload"], dict) else json.loads(job["payload"])
    transaction_id = payload.get("transaction_id")

    try:
        if not transaction_id:
            raise ValueError(f"job {job_id} payload does not contain transaction_id")

        transaction_data = await db.fetch_transaction(tenant_id, transaction_id)
        history_context = await context_builder.build_relevant_history(
            tenant_id,
            transaction_data["account_id"],
            transaction_data["counterparty_id"],
        )
        deterministic_assessment = rule_engine.evaluate(transaction_data, history_context)
        ai_summary = await context_builder.build_ai_summary(transaction_data, history_context, deterministic_assessment)
        final_record = await orchestrator_agent.process(
            job_id=job_id,
            tenant_id=tenant_id,
            transaction_id=transaction_id,
            deterministic=deterministic_assessment,
            context_summary=ai_summary,
        )
        assessment_id = await db.save_assessment(job_id, final_record)
        logger.info("Job %s completed, assessment=%s, status=%s", job_id, assessment_id, final_record.status)
    except Exception as exc:
        logger.exception("Failed to process job %s", job_id)
        await db.mark_job_failed(job_id, str(exc))


async def main_loop():
    db = Database()
    await db.connect()

    rule_engine = RuleEngine()
    context_builder = ContextBuilder(db)
    ai_client = AIClient()
    orchestrator_agent = AIOrchestrator(ai_client)

    logger.info("AML AI Worker started. AI configured=%s", ai_client.is_configured())

    try:
        while True:
            job = await db.claim_job()
            if job:
                logger.info("Claimed job %s", job["id"])
                await process_job(db, job, rule_engine, context_builder, orchestrator_agent)
            else:
                await asyncio.sleep(2)
    except asyncio.CancelledError:
        logger.info("Worker shutting down...")
    finally:
        await db.close()


if __name__ == "__main__":
    asyncio.run(main_loop())

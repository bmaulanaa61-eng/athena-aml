from models import AIEnrichmentOutput, RiskAssessmentRecord, DeterministicAssessment
from ai_client import AIClient, TransientAIError


class AIOrchestrator:
    def __init__(self, ai_client: AIClient):
        self.ai = ai_client

    async def process(
        self,
        job_id: str,
        tenant_id: str,
        transaction_id: str,
        deterministic: DeterministicAssessment,
        context_summary: dict,
    ) -> RiskAssessmentRecord:
        record = RiskAssessmentRecord(
            job_id=job_id,
            tenant_id=tenant_id,
            transaction_id=transaction_id,
            deterministic=deterministic,
            status="RULE_ONLY",
        )

        # Cost guardrail: LOW risk transactions do not require LLM enrichment.
        if deterministic.risk_level == "LOW":
            return record

        try:
            ai_raw_output = await self.ai.analyze_transaction(context_summary)
            record.ai_enrichment = AIEnrichmentOutput(**ai_raw_output)
            record.status = "AI_ENRICHED" if self.ai.is_configured() else "OFFLINE_ENRICHED"
            record.model_used = self.ai.model if self.ai.is_configured() else "offline-deterministic-fallback"
        except TransientAIError:
            record.status = "AI_FAILED"
        except Exception:
            record.status = "AI_FAILED"

        return record

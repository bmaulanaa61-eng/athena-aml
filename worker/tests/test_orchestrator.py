import pytest

from ai_client import AIClient
from models import DeterministicAssessment, RuleResult
from orchestrator import AIOrchestrator


@pytest.mark.asyncio
async def test_orchestrator_skips_ai_for_low_risk(monkeypatch):
    monkeypatch.delenv("AI_API_KEY", raising=False)
    orchestrator = AIOrchestrator(AIClient())

    record = await orchestrator.process(
        job_id="job-1",
        tenant_id="tenant-1",
        transaction_id="tx-1",
        deterministic=DeterministicAssessment(risk_score=0, risk_level="LOW", triggered_rules=[]),
        context_summary={},
    )

    assert record.status == "RULE_ONLY"
    assert record.ai_enrichment is None


@pytest.mark.asyncio
async def test_orchestrator_uses_offline_fallback_for_medium_risk(monkeypatch):
    monkeypatch.delenv("AI_API_KEY", raising=False)
    orchestrator = AIOrchestrator(AIClient())
    deterministic = DeterministicAssessment(
        risk_score=60,
        risk_level="MEDIUM",
        triggered_rules=[RuleResult(rule_code="SML-0002", triggered=True, risk_weight=25, signal="Velocity")],
    )

    record = await orchestrator.process(
        job_id="job-1",
        tenant_id="tenant-1",
        transaction_id="tx-1",
        deterministic=deterministic,
        context_summary={"deterministic_risk_assessment": {"score": 60, "level": "MEDIUM", "triggered_rules": [{"signal": "Velocity"}]}},
    )

    assert record.status == "OFFLINE_ENRICHED"
    assert record.ai_enrichment is not None
    assert record.ai_enrichment.assessment == "REVIEW"

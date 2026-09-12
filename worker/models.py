from typing import List, Optional
from pydantic import BaseModel, Field


class RuleResult(BaseModel):
    rule_code: str
    triggered: bool
    risk_weight: int
    signal: str


class DeterministicAssessment(BaseModel):
    risk_score: int = Field(ge=0, le=100)
    risk_level: str
    triggered_rules: List[RuleResult]


class AIEnrichmentOutput(BaseModel):
    assessment: str = Field(description="REVIEW, APPROVE, ESCALATE, or CLEAR")
    confidence_score: float = Field(ge=0.0, le=1.0)
    risk_factors: List[str] = Field(default_factory=list)
    supporting_evidence: List[str] = Field(default_factory=list)
    investigation_suggestions: List[str] = Field(default_factory=list)
    explanation: str = ""


class RiskAssessmentRecord(BaseModel):
    job_id: str
    tenant_id: str
    transaction_id: str
    deterministic: DeterministicAssessment
    ai_enrichment: Optional[AIEnrichmentOutput] = None
    status: str
    model_used: Optional[str] = None


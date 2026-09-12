import os
import json
from openai import AsyncOpenAI
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type


class TransientAIError(Exception):
    pass


class AIClient:
    def __init__(self):
        self.api_key = os.getenv("AI_API_KEY", "")
        self.client = AsyncOpenAI(
            api_key=self.api_key or "offline-mode",
            base_url=os.getenv("AI_BASE_URL", "https://generativelanguage.googleapis.com/v1beta/openai/"),
        )
        self.model = os.getenv("AI_MODEL_NAME", "gemini-3.5-flash-lite")

    def is_configured(self) -> bool:
        return bool(self.api_key)

    @retry(
        stop=stop_after_attempt(3),
        wait=wait_exponential(multiplier=1, min=2, max=10),
        retry=retry_if_exception_type(TransientAIError),
        reraise=True,
    )
    async def analyze_transaction(self, context_summary: dict) -> dict:
        if not self.is_configured():
            return self._offline_fallback(context_summary)

        try:
            response = await self.client.chat.completions.create(
                model=self.model,
                response_format={"type": "json_object"},
                messages=[
                    {
                        "role": "system",
                        "content": (
                            "You are an AML investigation assistant. Return strict JSON with: "
                            "assessment (APPROVE|REVIEW|ESCALATE), confidence_score (0-1), "
                            "risk_factors, supporting_evidence, investigation_suggestions, explanation."
                        ),
                    },
                    {"role": "user", "content": json.dumps(context_summary, default=str)},
                ],
                temperature=0.2,
            )
            raw_response = response.choices[0].message.content
            return json.loads(raw_response)
        except Exception as exc:
            raise TransientAIError(f"AI API request failed: {exc}") from exc

    def _offline_fallback(self, context_summary: dict) -> dict:
        risk = context_summary.get("deterministic_risk_assessment", {})
        level = risk.get("level", "LOW")
        score = int(risk.get("score", 0))
        triggered = risk.get("triggered_rules", [])
        assessment = "APPROVE" if level == "LOW" else "REVIEW"
        if score >= 85:
            assessment = "ESCALATE"
        return {
            "assessment": assessment,
            "confidence_score": 0.72,
            "risk_factors": [rule.get("signal", rule.get("rule_code", "rule")) for rule in triggered],
            "supporting_evidence": [f"Deterministic AML score={score}, level={level}"],
            "investigation_suggestions": [
                "Review account and counterparty relationship.",
                "Compare recent transaction pattern against customer baseline.",
            ],
            "explanation": "Offline deterministic fallback used because AI_API_KEY is not configured.",
        }

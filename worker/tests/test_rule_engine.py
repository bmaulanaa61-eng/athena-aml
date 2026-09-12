import pytest

from rule_engine import RuleEngine


def test_rule_engine_flags_high_risk_transaction():
    engine = RuleEngine()
    assessment = engine.evaluate(
        {"amount": 1000, "occurred_at": "2026-09-12T02:30:00Z"},
        {
            "recent_transactions": [
                {"amount": 100},
                {"amount": 120},
                {"amount": 90},
                {"amount": 110},
            ],
            "is_new_counterparty": True,
            "is_off_hours": True,
        },
    )

    assert assessment.risk_score == 100
    assert assessment.risk_level == "HIGH"
    assert {rule.rule_code for rule in assessment.triggered_rules} == {"SML-0001", "SML-0002", "SML-0003", "SML-0004"}


def test_rule_engine_low_risk_when_no_signal():
    engine = RuleEngine()
    assessment = engine.evaluate(
        {"amount": 100},
        {"recent_transactions": [{"amount": 100}, {"amount": 110}], "is_new_counterparty": False},
    )

    assert assessment.risk_score == 0
    assert assessment.risk_level == "LOW"
    assert assessment.triggered_rules == []

from datetime import time
from typing import List
from models import RuleResult, DeterministicAssessment

class RuleEngine:
    def __init__(self):
        pass

    def evaluate(self, transaction_data: dict, history_context: dict) -> DeterministicAssessment:
        triggered_rules: List[RuleResult] = []
        total_risk_score = 0
        
        # Rule 1: Pattern-based Amount Anomaly (Standard Weight: 35)
        # Transaksi > 300% dari rata-rata histori pengguna mengindikasikan anomali lonjakan nilai masif.
        amount = float(transaction_data.get('amount', 0.0))
        recent_txs = history_context.get("recent_transactions", [])
        if recent_txs:
            avg_amount = sum(float(tx['amount']) for tx in recent_txs) / len(recent_txs)
            if amount > (avg_amount * 3):
                triggered_rules.append(RuleResult(
                    rule_code="SML-0001",
                    triggered=True,
                    risk_weight=35,
                    signal="Unusual Amount Spike (>300% Historical Average)"
                ))
                total_risk_score += 35

        # Rule 2: Velocity & Rapid Frequency (Standard Weight: 25)
        # > 3 transaksi dalam jendela waktu singkat memicu indikasi structuring/smurfing.
        recent_tx_count = len(history_context.get("recent_transactions", []))
        if recent_tx_count > 3:
            triggered_rules.append(RuleResult(
                rule_code="SML-0002",
                triggered=True,
                risk_weight=25,
                signal="High Frequency Velocity (Potential Structuring)"
            ))
            total_risk_score += 25
            
        # Rule 3: New / Unverified Counterparty (Standard Weight: 25)
        # Transfer ke entitas baru yang belum pernah berinteraksi dalam riwayat 30 hari.
        is_new = history_context.get('is_new_counterparty', False)
        if is_new:
            triggered_rules.append(RuleResult(
                rule_code="SML-0003", 
                triggered=True, 
                risk_weight=25, 
                signal="New Unverified Counterparty Transfer"
            ))
            total_risk_score += 25

        # Rule 4: Temporal / Off-Hours Behavioral Anomaly (Standard Weight: 15)
        # Transaksi pada jam dorman/dini hari (01:00 - 05:00) di luar kebiasaan jam bangun nasabah.
        is_off_hours = history_context.get('is_off_hours', False)
        occurred_at = transaction_data.get('occurred_at')
        if not is_off_hours and hasattr(occurred_at, 'time'):
            is_off_hours = time(1, 0) <= occurred_at.time() <= time(5, 0)
        if is_off_hours:
            triggered_rules.append(RuleResult(
                rule_code="SML-0004",
                triggered=True,
                risk_weight=15,
                signal="Off-Hours Transaction (01:00 - 05:00)"
            ))
            total_risk_score += 15

        total_risk_score = min(total_risk_score, 100)

        if total_risk_score >= 70:
            risk_level = "HIGH"
        elif total_risk_score >= 30:
            risk_level = "MEDIUM"
        else:
            risk_level = "LOW"

        return DeterministicAssessment(
            risk_score=total_risk_score,
            risk_level=risk_level,
            triggered_rules=triggered_rules
        )
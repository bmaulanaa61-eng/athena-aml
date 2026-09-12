from datetime import time


class ContextBuilder:
    def __init__(self, db_client):
        self.db = db_client

    async def build_relevant_history(self, tenant_id: str, account_id: str, counterparty_id: str) -> dict:
        async with self.db.conn.cursor() as cur:
            query = """
                SELECT id, amount, currency, transaction_type, occurred_at
                FROM transactions
                WHERE tenant_id = %s::uuid AND (account_id = %s OR counterparty_id = %s)
                  AND occurred_at >= NOW() - INTERVAL '30 days'
                ORDER BY occurred_at DESC
                LIMIT 100;
            """
            await cur.execute(query, (tenant_id, account_id, counterparty_id))
            recent_txs = await cur.fetchall()

            check_new_query = """
                SELECT 1 FROM transactions
                WHERE tenant_id = %s::uuid AND account_id = %s AND counterparty_id = %s
                  AND occurred_at < NOW() - INTERVAL '30 days'
                LIMIT 1;
            """
            await cur.execute(check_new_query, (tenant_id, account_id, counterparty_id))
            is_new = (await cur.fetchone()) is None

            return {
                "recent_transactions": recent_txs,
                "is_new_counterparty": is_new,
            }

    async def build_ai_summary(self, transaction_data: dict, history_context: dict, deterministic_assessment) -> dict:
        occurred_at = transaction_data.get("occurred_at")
        if hasattr(occurred_at, "time"):
            history_context["is_off_hours"] = time(1, 0) <= occurred_at.time() <= time(5, 0)

        return {
            "current_transaction": {
                "amount": transaction_data.get("amount"),
                "currency": transaction_data.get("currency"),
                "transaction_type": transaction_data.get("transaction_type"),
                "occurred_at": str(occurred_at) if occurred_at is not None else None,
            },
            "relevant_history_summary": {
                "transaction_count_last_30_days": len(history_context.get("recent_transactions", [])),
                "is_new_counterparty": history_context.get("is_new_counterparty", False),
                "is_off_hours": history_context.get("is_off_hours", False),
            },
            "deterministic_risk_assessment": {
                "score": deterministic_assessment.risk_score,
                "level": deterministic_assessment.risk_level,
                "triggered_rules": [rule.model_dump() for rule in deterministic_assessment.triggered_rules],
            },
        }

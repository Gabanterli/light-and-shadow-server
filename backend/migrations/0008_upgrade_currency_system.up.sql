-- Migration 0008: Atualizar colunas de moeda para suportar uint64 (BIGINT)
ALTER TABLE characters ALTER COLUMN gold TYPE BIGINT;

ALTER TABLE market_orders 
    ALTER COLUMN price_gold TYPE BIGINT,
    ALTER COLUMN tax_gold TYPE BIGINT;

ALTER TABLE market_history 
    ALTER COLUMN price_gold TYPE BIGINT,
    ALTER COLUMN tax_gold TYPE BIGINT;

ALTER TABLE trade_logs 
    ALTER COLUMN gold_a TYPE BIGINT,
    ALTER COLUMN gold_b TYPE BIGINT;

-- Drop indexes
DROP INDEX IF EXISTS idx_crafting_recipes_profession;
DROP INDEX IF EXISTS idx_crafting_recipes_result_item;

DROP INDEX IF EXISTS idx_trade_items_owner_id;
DROP INDEX IF EXISTS idx_trade_items_trade_id;

DROP INDEX IF EXISTS idx_trades_expires_at;
DROP INDEX IF EXISTS idx_trades_status;
DROP INDEX IF EXISTS idx_trades_recipient_id;
DROP INDEX IF EXISTS idx_trades_initiator_id;

DROP INDEX IF EXISTS idx_equipment_slot;
DROP INDEX IF EXISTS idx_equipment_character_id;

DROP INDEX IF EXISTS idx_inventory_items_slot;
DROP INDEX IF EXISTS idx_inventory_items_item_id;
DROP INDEX IF EXISTS idx_inventory_items_character_id;

DROP INDEX IF EXISTS idx_items_name;
DROP INDEX IF EXISTS idx_items_level;
DROP INDEX IF EXISTS idx_items_rarity;
DROP INDEX IF EXISTS idx_items_type;

-- Drop tables in reverse order of creation (respecting foreign key dependencies)
DROP TABLE IF EXISTS crafting_recipes;
DROP TABLE IF EXISTS trade_items;
DROP TABLE IF EXISTS trades;
DROP TABLE IF EXISTS equipment;
DROP TABLE IF EXISTS inventory_items;
DROP TABLE IF EXISTS inventories;
DROP TABLE IF EXISTS items; 
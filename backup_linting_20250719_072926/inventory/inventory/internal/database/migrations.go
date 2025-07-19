package database

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// RunMigrations exécute les migrations de base de données
func RunMigrations(db *DB) error {
	logrus.Info("Running database migrations...")

	migrations := []string{
		createItemsTable,
		createInventoriesTable,
		createEquipmentTable,
		createTradesTable,
		createCraftingRecipesTable,
		createIndexes,
	}

	for i, migration := range migrations {
		logrus.WithField("migration", i+1).Debug("Executing migration")

		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration %d: %w", i+1, err)
		}
	}

	logrus.Info("Database migrations completed successfully")
	return nil
}

// Définition des migrations SQL
const createItemsTable = `
CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    item_type VARCHAR(100) NOT NULL,
    rarity VARCHAR(50) NOT NULL DEFAULT 'common',
    stackable BOOLEAN NOT NULL DEFAULT false,
    max_stack INTEGER DEFAULT 1,
    tradable BOOLEAN NOT NULL DEFAULT true,
    value INTEGER NOT NULL DEFAULT 0,
    stats JSONB,
    requirements JSONB,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const createInventoriesTable = `
CREATE TABLE IF NOT EXISTS inventories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL UNIQUE,
    slots JSONB NOT NULL DEFAULT '[]',
    max_slots INTEGER NOT NULL DEFAULT 50,
    gold INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const createEquipmentTable = `
CREATE TABLE IF NOT EXISTS equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    character_id UUID NOT NULL UNIQUE,
    head_slot UUID REFERENCES items(id),
    chest_slot UUID REFERENCES items(id),
    legs_slot UUID REFERENCES items(id),
    feet_slot UUID REFERENCES items(id),
    main_hand_slot UUID REFERENCES items(id),
    off_hand_slot UUID REFERENCES items(id),
    ring1_slot UUID REFERENCES items(id),
    ring2_slot UUID REFERENCES items(id),
    necklace_slot UUID REFERENCES items(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const createTradesTable = `
CREATE TABLE IF NOT EXISTS trades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_character_id UUID NOT NULL,
    to_character_id UUID NOT NULL,
    from_items JSONB,
    to_items JSONB,
    from_gold INTEGER DEFAULT 0,
    to_gold INTEGER DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    from_accepted BOOLEAN DEFAULT false,
    to_accepted BOOLEAN DEFAULT false,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const createCraftingRecipesTable = `
CREATE TABLE IF NOT EXISTS crafting_recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    result_item_id UUID NOT NULL REFERENCES items(id),
    result_quantity INTEGER NOT NULL DEFAULT 1,
    required_items JSONB NOT NULL,
    required_skill_level INTEGER DEFAULT 0,
    crafting_time INTEGER DEFAULT 0,
    success_rate FLOAT DEFAULT 1.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const createIndexes = `
CREATE INDEX IF NOT EXISTS idx_inventories_character_id ON inventories(character_id);
CREATE INDEX IF NOT EXISTS idx_equipment_character_id ON equipment(character_id);
CREATE INDEX IF NOT EXISTS idx_trades_from_character ON trades(from_character_id);
CREATE INDEX IF NOT EXISTS idx_trades_to_character ON trades(to_character_id);
CREATE INDEX IF NOT EXISTS idx_trades_status ON trades(status);
CREATE INDEX IF NOT EXISTS idx_items_type ON items(item_type);
CREATE INDEX IF NOT EXISTS idx_items_rarity ON items(rarity);
CREATE INDEX IF NOT EXISTS idx_crafting_recipes_result_item ON crafting_recipes(result_item_id);
`

-- Create items table
CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    rarity VARCHAR(50) NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    stats JSONB,
    requirements JSONB,
    metadata JSONB,
    weight DECIMAL(10,3) NOT NULL DEFAULT 0.0,
    max_stack_size INTEGER NOT NULL DEFAULT 1,
    tradeable BOOLEAN NOT NULL DEFAULT true,
    sellable BOOLEAN NOT NULL DEFAULT true,
    destroyable BOOLEAN NOT NULL DEFAULT true,
    value BIGINT NOT NULL DEFAULT 0,
    image_url VARCHAR(512),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create inventories table
CREATE TABLE IF NOT EXISTS inventories (
    character_id UUID PRIMARY KEY,
    slots INTEGER NOT NULL DEFAULT 30,
    max_weight DECIMAL(10,3) NOT NULL DEFAULT 100.0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create inventory_items table
CREATE TABLE IF NOT EXISTS inventory_items (
    id UUID PRIMARY KEY,
    character_id UUID NOT NULL REFERENCES inventories(character_id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 1,
    slot INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(character_id, slot),
    CONSTRAINT positive_quantity CHECK (quantity > 0),
    CONSTRAINT valid_slot CHECK (slot >= 0)
);

-- Create equipment table
CREATE TABLE IF NOT EXISTS equipment (
    id UUID PRIMARY KEY,
    character_id UUID NOT NULL,
    slot VARCHAR(50) NOT NULL,
    item_id UUID REFERENCES items(id) ON DELETE SET NULL,
    equipped_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(character_id, slot)
);

-- Create trades table
CREATE TABLE IF NOT EXISTS trades (
    id UUID PRIMARY KEY,
    initiator_id UUID NOT NULL,
    recipient_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    initiator_ready BOOLEAN NOT NULL DEFAULT false,
    recipient_ready BOOLEAN NOT NULL DEFAULT false,
    initiator_gold BIGINT NOT NULL DEFAULT 0,
    recipient_gold BIGINT NOT NULL DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT different_players CHECK (initiator_id != recipient_id),
    CONSTRAINT non_negative_gold CHECK (initiator_gold >= 0 AND recipient_gold >= 0)
);

-- Create trade_items table
CREATE TABLE IF NOT EXISTS trade_items (
    id UUID PRIMARY KEY,
    trade_id UUID NOT NULL REFERENCES trades(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL,
    owner_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT positive_quantity CHECK (quantity > 0)
);

-- Create crafting_recipes table
CREATE TABLE IF NOT EXISTS crafting_recipes (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    result_item_id UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    result_quantity INTEGER NOT NULL DEFAULT 1,
    required_level INTEGER NOT NULL DEFAULT 1,
    required_profession VARCHAR(100),
    materials JSONB NOT NULL,
    crafting_time INTEGER NOT NULL DEFAULT 1000, -- in milliseconds
    experience_gain INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT positive_result_quantity CHECK (result_quantity > 0),
    CONSTRAINT positive_crafting_time CHECK (crafting_time > 0)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_items_type ON items(type);
CREATE INDEX IF NOT EXISTS idx_items_rarity ON items(rarity);
CREATE INDEX IF NOT EXISTS idx_items_level ON items(level);
CREATE INDEX IF NOT EXISTS idx_items_name ON items(name);

CREATE INDEX IF NOT EXISTS idx_inventory_items_character_id ON inventory_items(character_id);
CREATE INDEX IF NOT EXISTS idx_inventory_items_item_id ON inventory_items(item_id);
CREATE INDEX IF NOT EXISTS idx_inventory_items_slot ON inventory_items(character_id, slot);

CREATE INDEX IF NOT EXISTS idx_equipment_character_id ON equipment(character_id);
CREATE INDEX IF NOT EXISTS idx_equipment_slot ON equipment(character_id, slot);

CREATE INDEX IF NOT EXISTS idx_trades_initiator_id ON trades(initiator_id);
CREATE INDEX IF NOT EXISTS idx_trades_recipient_id ON trades(recipient_id);
CREATE INDEX IF NOT EXISTS idx_trades_status ON trades(status);
CREATE INDEX IF NOT EXISTS idx_trades_expires_at ON trades(expires_at);

CREATE INDEX IF NOT EXISTS idx_trade_items_trade_id ON trade_items(trade_id);
CREATE INDEX IF NOT EXISTS idx_trade_items_owner_id ON trade_items(owner_id);

CREATE INDEX IF NOT EXISTS idx_crafting_recipes_result_item ON crafting_recipes(result_item_id);
CREATE INDEX IF NOT EXISTS idx_crafting_recipes_profession ON crafting_recipes(required_profession); 
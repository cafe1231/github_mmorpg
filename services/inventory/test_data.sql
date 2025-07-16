-- Données de test pour le service Inventory
-- À exécuter APRÈS que les migrations aient créé les tables

-- Insérer des items de test
INSERT INTO items (id, name, description, item_type, rarity, stackable, max_stack, tradable, value, created_at, updated_at) VALUES 
(
    '123e4567-e89b-12d3-a456-426614174000',
    'Épée de fer',
    'Une épée solide en fer forgé',
    'weapon',
    'common',
    false,
    1,
    true,
    150,
    NOW(),
    NOW()
),
(
    '456e7890-e89b-12d3-a456-426614174001',
    'Potion de santé',
    'Restaure 50 points de vie',
    'consumable',
    'common',
    true,
    99,
    true,
    25,
    NOW(),
    NOW()
),
(
    '789abcde-e89b-12d3-a456-426614174002',
    'Casque de cuir',
    'Protection basique pour la tête',
    'armor',
    'common',
    false,
    1,
    true,
    75,
    NOW(),
    NOW()
),
(
    '987fcdeb-e89b-12d3-a456-426614174003',
    'Minerai de fer',
    'Matériau de base pour la forge',
    'material',
    'common',
    true,
    50,
    true,
    10,
    NOW(),
    NOW()
),
(
    '654321ab-e89b-12d3-a456-426614174004',
    'Anneau magique',
    'Augmente la régénération de mana',
    'misc',
    'rare',
    false,
    1,
    true,
    500,
    NOW(),
    NOW()
);

-- Créer un inventaire de test pour le personnage de test
INSERT INTO inventories (id, character_id, max_slots, gold, created_at, updated_at) VALUES 
(
    '550e8400-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440000',
    50,
    1000,
    NOW(),
    NOW()
);

-- Ajouter quelques items dans l'inventaire de test
INSERT INTO inventory_items (id, character_id, item_id, quantity, slot, created_at, updated_at) VALUES 
(
    '111e1111-e29b-41d4-a716-446655440001',
    '550e8400-e29b-41d4-a716-446655440000',
    '456e7890-e89b-12d3-a456-426614174001',
    5,
    0,
    NOW(),
    NOW()
),
(
    '222e2222-e29b-41d4-a716-446655440002',
    '550e8400-e29b-41d4-a716-446655440000',
    '987fcdeb-e89b-12d3-a456-426614174003',
    10,
    1,
    NOW(),
    NOW()
); 
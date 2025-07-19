-- Script de vérification des tables créées par les migrations
-- À exécuter dans inventory_db

-- 1. Lister toutes les tables dans le schéma public
SELECT table_name, table_type 
FROM information_schema.tables 
WHERE table_schema = 'public' 
ORDER BY table_name;

-- 2. Vérifier la structure de chaque table
\d items;
\d inventories;
\d equipment;
\d trades;
\d crafting_recipes;

-- 3. Compter les enregistrements dans chaque table
SELECT 'items' as table_name, COUNT(*) as count FROM items
UNION ALL
SELECT 'inventories' as table_name, COUNT(*) as count FROM inventories
UNION ALL
SELECT 'equipment' as table_name, COUNT(*) as count FROM equipment
UNION ALL
SELECT 'trades' as table_name, COUNT(*) as count FROM trades
UNION ALL
SELECT 'crafting_recipes' as table_name, COUNT(*) as count FROM crafting_recipes;

-- 4. Vérifier les indexes créés
SELECT indexname, tablename 
FROM pg_indexes 
WHERE schemaname = 'public'
ORDER BY tablename, indexname; 
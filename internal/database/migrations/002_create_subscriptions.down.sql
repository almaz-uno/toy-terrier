-- Drop subscriptions and categories tables
DROP TRIGGER IF EXISTS update_subscriptions_updated_at ON subscriptions;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS categories;

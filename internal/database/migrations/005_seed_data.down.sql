-- Remove seed data
DELETE FROM categories WHERE slug IN ('hot-toys', 'threezero', 'sideshow', 'facebook-hot-toys');

-- Drop additional indexes
DROP INDEX IF EXISTS idx_notifications_pending;
DROP INDEX IF EXISTS idx_topics_unnotified;
DROP INDEX IF EXISTS idx_users_active_subscriptions;

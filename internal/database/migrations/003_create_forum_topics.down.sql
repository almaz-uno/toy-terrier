-- Drop forum_topics table
DROP TRIGGER IF EXISTS update_forum_topics_updated_at ON forum_topics;
DROP TABLE IF EXISTS forum_topics;

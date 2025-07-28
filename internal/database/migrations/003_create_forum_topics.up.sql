-- Create forum_topics table
CREATE TABLE IF NOT EXISTS forum_topics (
    id SERIAL PRIMARY KEY,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    author VARCHAR(255) NOT NULL,
    source_topic_id VARCHAR(100) UNIQUE NOT NULL,
    source_url VARCHAR(500) UNIQUE NOT NULL,
    content TEXT NOT NULL,
    reply_count INTEGER NOT NULL DEFAULT 0,
    view_count INTEGER NOT NULL DEFAULT 0,
    last_post_time TIMESTAMP WITH TIME ZONE NOT NULL,
    content_hash VARCHAR(64) UNIQUE NOT NULL,
    has_images BOOLEAN NOT NULL DEFAULT false,
    image_urls TEXT[],
    is_notified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_topics_category_id ON forum_topics(category_id);
CREATE INDEX IF NOT EXISTS idx_topics_source_topic_id ON forum_topics(source_topic_id);
CREATE INDEX IF NOT EXISTS idx_topics_content_hash ON forum_topics(content_hash);
CREATE INDEX IF NOT EXISTS idx_topics_created_at ON forum_topics(created_at);
CREATE INDEX IF NOT EXISTS idx_topics_is_notified ON forum_topics(is_notified);
CREATE INDEX IF NOT EXISTS idx_topics_last_post_time ON forum_topics(last_post_time);

-- Create trigger for updated_at
CREATE TRIGGER update_forum_topics_updated_at
    BEFORE UPDATE ON forum_topics
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

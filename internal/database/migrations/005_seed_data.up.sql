-- Insert default categories
INSERT INTO categories (name, slug, source_type, source_url, is_active) VALUES
('Hot Toys', 'hot-toys', 'forum', 'https://bbs.bbicn.com/forumdisplay.php?fid=20', true),
('ThreeZero', 'threezero', 'forum', 'https://bbs.bbicn.com/forumdisplay.php?fid=21', true),
('Sideshow', 'sideshow', 'forum', 'https://bbs.bbicn.com/forumdisplay.php?fid=22', true),
('Facebook Hot Toys', 'facebook-hot-toys', 'facebook', 'https://www.facebook.com/hottoys', true)
ON CONFLICT (slug) DO NOTHING;

-- Create additional indexes for performance
CREATE INDEX IF NOT EXISTS idx_notifications_pending ON notifications(status, next_retry_at)
    WHERE status IN ('pending', 'rate_limited');

CREATE INDEX IF NOT EXISTS idx_topics_unnotified ON forum_topics(category_id, created_at)
    WHERE is_notified = false;

CREATE INDEX IF NOT EXISTS idx_users_active_subscriptions ON subscriptions(user_id, is_active)
    WHERE is_active = true;

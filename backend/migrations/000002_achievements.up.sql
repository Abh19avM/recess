-- Create Achievements Table
CREATE TABLE IF NOT EXISTS achievements (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(32) NOT NULL,
    icon VARCHAR(32) NOT NULL DEFAULT '🎖️',
    badge_tone VARCHAR(32) NOT NULL DEFAULT 'blue',
    points INTEGER NOT NULL DEFAULT 10,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create User Achievements Table
CREATE TABLE IF NOT EXISTS user_achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id VARCHAR(64) NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_achievement UNIQUE (user_id, achievement_id)
);

CREATE INDEX IF NOT EXISTS idx_user_achievements_user ON user_achievements(user_id);
CREATE INDEX IF NOT EXISTS idx_user_achievements_unlocked ON user_achievements(unlocked_at DESC);

-- Seed Standard Recess School Playground Achievements
INSERT INTO achievements (id, name, description, category, icon, badge_tone, points)
VALUES
    ('first_win', 'First Bell Victory', 'Win your very first classroom duel in Recess', 'general', '🔔', 'green', 10),
    ('streak_3', 'Hat-Trick Student', 'Win 3 matches in a row across any game', 'general', '🔥', 'red', 25),
    ('streak_5', 'Classroom Dominator', 'Achieve an unbroken 5-match winning streak', 'general', '⚡', 'purple', 50),
    ('cricket_centurion', 'Finger Cricket Master', 'Score 50+ runs in a single Hand Cricket innings', 'hand_cricket', '🏏', 'amber', 30),
    ('clean_sweep_xo', 'Flawless Grid', 'Win an XO duel without allowing opponent a corner', 'xo', '✕', 'blue', 20),
    ('box_conqueror', 'Territory Mogul', 'Claim 6 or more boxes in a single Dots & Boxes duel', 'dots_boxes', '⚄', 'green', 25),
    ('scholar_1300', 'Honor Roll ELO', 'Reach an overall rating of 1300 ELO', 'rating', '📜', 'amber', 50),
    ('veteran_25', 'Recess Veteran', 'Complete 25 total multiplayer classroom matches', 'milestone', '🎓', 'purple', 40)
ON CONFLICT (id) DO NOTHING;

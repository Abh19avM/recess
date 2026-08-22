-- name: UpsertRating :one
INSERT INTO ratings (
    user_id,
    game_type,
    rating,
    wins,
    losses,
    draws,
    win_streak,
    best_win_streak,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, NOW()
)
ON CONFLICT (user_id, game_type) DO UPDATE
SET rating = EXCLUDED.rating,
    wins = EXCLUDED.wins,
    losses = EXCLUDED.losses,
    draws = EXCLUDED.draws,
    win_streak = EXCLUDED.win_streak,
    best_win_streak = GREATEST(ratings.best_win_streak, EXCLUDED.best_win_streak),
    updated_at = NOW()
RETURNING id, user_id, game_type, rating, wins, losses, draws, win_streak, best_win_streak, updated_at;

-- name: GetUserRating :one
SELECT id, user_id, game_type, rating, wins, losses, draws, win_streak, best_win_streak, updated_at
FROM ratings
WHERE user_id = $1 AND game_type = $2 LIMIT 1;

-- name: GetUserAllRatings :many
SELECT id, user_id, game_type, rating, wins, losses, draws, win_streak, best_win_streak, updated_at
FROM ratings
WHERE user_id = $1
ORDER BY game_type ASC;

-- name: GetLeaderboardByGame :many
SELECT r.id, r.user_id, r.game_type, r.rating, r.wins, r.losses, r.draws, r.win_streak, r.best_win_streak,
       u.username, p.nickname, p.avatar_preset, p.title
FROM ratings r
JOIN users u ON r.user_id = u.id
JOIN profiles p ON r.user_id = p.user_id
WHERE r.game_type = $1
ORDER BY r.rating DESC, r.wins DESC
LIMIT $2;

-- name: CreateMatch :one
INSERT INTO matches (
    room_id,
    game_type,
    status,
    metadata
) VALUES (
    $1, $2, $3, $4
)
RETURNING id, room_id, game_type, status, winner_id, started_at, ended_at, duration_seconds, metadata;

-- name: GetMatchByID :one
SELECT id, room_id, game_type, status, winner_id, started_at, ended_at, duration_seconds, metadata
FROM matches
WHERE id = $1 LIMIT 1;

-- name: CompleteMatch :one
UPDATE matches
SET status = 'completed',
    winner_id = $2,
    ended_at = NOW(),
    duration_seconds = $3,
    metadata = $4
WHERE id = $1
RETURNING id, room_id, game_type, status, winner_id, started_at, ended_at, duration_seconds, metadata;

-- name: AddMatchPlayer :one
INSERT INTO match_players (
    match_id,
    user_id,
    player_index,
    score,
    is_winner,
    rating_before,
    rating_after
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, match_id, user_id, player_index, score, is_winner, rating_before, rating_after, created_at;

-- name: GetMatchPlayers :many
SELECT mp.id, mp.match_id, mp.user_id, mp.player_index, mp.score, mp.is_winner, mp.rating_before, mp.rating_after, mp.created_at,
       u.username, p.avatar_preset
FROM match_players mp
JOIN users u ON mp.user_id = u.id
JOIN profiles p ON mp.user_id = p.user_id
WHERE mp.match_id = $1
ORDER BY mp.player_index ASC;

-- name: CreateGameResult :one
INSERT INTO game_results (
    match_id,
    game_type,
    rounds_played,
    summary
) VALUES (
    $1, $2, $3, $4
)
RETURNING id, match_id, game_type, rounds_played, summary, created_at;

-- name: GetGameResultByMatchID :one
SELECT id, match_id, game_type, rounds_played, summary, created_at
FROM game_results
WHERE match_id = $1 LIMIT 1;

-- name: CreateRoom :one
INSERT INTO rooms (
    code,
    game_type,
    title,
    host_id,
    status,
    max_players,
    is_private,
    passcode
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id, code, game_type, title, host_id, status, max_players, is_private, passcode, created_at, updated_at;

-- name: GetRoomByCode :one
SELECT r.id, r.code, r.game_type, r.title, r.host_id, r.status, r.max_players, r.is_private, r.passcode, r.created_at, r.updated_at,
       u.username AS host_username, p.avatar_preset AS host_avatar
FROM rooms r
JOIN users u ON r.host_id = u.id
JOIN profiles p ON r.host_id = p.user_id
WHERE r.code = $1 LIMIT 1;

-- name: ListActiveRooms :many
SELECT r.id, r.code, r.game_type, r.title, r.host_id, r.status, r.max_players, r.is_private, r.created_at, r.updated_at,
       u.username AS host_username, p.avatar_preset AS host_avatar
FROM rooms r
JOIN users u ON r.host_id = u.id
JOIN profiles p ON r.host_id = p.user_id
WHERE r.is_private = FALSE AND r.status != 'completed' AND r.status != 'cancelled'
ORDER BY r.created_at DESC
LIMIT 50;

-- name: UpdateRoomStatus :one
UPDATE rooms
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, code, game_type, title, host_id, status, max_players, is_private, updated_at;

-- name: DeleteRoom :exec
DELETE FROM rooms
WHERE code = $1;

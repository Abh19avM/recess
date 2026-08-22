-- name: CreateUser :one
INSERT INTO users (
    username,
    email,
    password_hash,
    is_guest
) VALUES (
    $1, $2, $3, $4
)
RETURNING id, username, email, is_guest, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, username, email, is_guest, created_at, updated_at
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT id, username, email, password_hash, is_guest, created_at, updated_at
FROM users
WHERE username = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, username, email, password_hash, is_guest, created_at, updated_at
FROM users
WHERE email = $1 LIMIT 1;

-- name: CreateProfile :one
INSERT INTO profiles (
    user_id,
    nickname,
    avatar_preset,
    title,
    bio
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, user_id, nickname, avatar_preset, title, bio, created_at, updated_at;

-- name: GetProfileByUserID :one
SELECT p.id, p.user_id, p.nickname, p.avatar_preset, p.title, p.bio, p.created_at, p.updated_at,
       u.username, u.email, u.is_guest
FROM profiles p
JOIN users u ON p.user_id = u.id
WHERE p.user_id = $1 LIMIT 1;

-- name: UpdateProfile :one
UPDATE profiles
SET nickname = $2,
    avatar_preset = $3,
    title = $4,
    bio = $5,
    updated_at = NOW()
WHERE user_id = $1
RETURNING id, user_id, nickname, avatar_preset, title, bio, created_at, updated_at;

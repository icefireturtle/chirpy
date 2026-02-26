-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (gen_random_uuid(), now(), now(), $1, $2)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email=$1;

-- name: GetUserFromRefreshToken :one
SELECT users.* FROM users
INNER JOIN refresh_tokens ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token=$1 
AND refresh_tokens.revoked_at IS NULL 
AND refresh_tokens.expires_at > now();

-- name: UpdateUser :one
UPDATE users
SET email=$2, hashed_password=$3, updated_at=now()
WHERE id=$1
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: IsChirpyRed :exec
UPDATE users
SET is_chirpy_red = true
WHERE id=$1;
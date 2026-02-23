-- name: CreateChirps :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (gen_random_uuid(), now(), now(), $1, $2)
RETURNING *;

-- name: ViewChirps :many

SELECT *
FROM chirps
ORDER BY created_at ASC;

-- name: ViewChirp :one

SELECT *
FROM chirps
WHERE id=$1;
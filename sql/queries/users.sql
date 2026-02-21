-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (gen_rand_uuid(), now(), now(), $1)
RETURNING *;
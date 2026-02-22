-- name: ViewChirp :one

SELECT *
FROM chirps
WHERE id=$1;
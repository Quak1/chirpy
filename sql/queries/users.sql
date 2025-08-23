-- name: CreateUser :one
INSERT INTO users (hashed_password, email)
VALUES ($1, $2)
RETURNING *;


-- name: DeleteUsers :exec
DELETE FROM users;


-- name: GetUserByEmail :one
SELECT * FROM users
WHERE users.email = $1;


-- name: UpdateUser :one
UPDATE users
SET hashed_password = $2, email = $3, updated_at = $4
WHERE users.id = $1
RETURNING *;


-- name: UpgradeChirpyRed :exec
UPDATE users
SET is_chirpy_red = true
WHERE id = $1;

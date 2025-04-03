-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: GetUserFromEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUserEmailPassword :one
UPDATE users
SET email = $1, hashed_password = $2, updated_at = NOW()
WHERE id = $3
RETURNING *;

-- name: UpgradeUser :exec
UPDATE users
SET is_chirpy_red = true
WHERE ID = $1;
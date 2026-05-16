-- name: CreateAddress :exec
INSERT INTO addresses (id, user_id, street, city, state, zip_code, country, is_default) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListAddressesByUserID :many
SELECT * FROM addresses WHERE user_id=$1;

-- name: UpdateAddress :exec
UPDATE addresses SET street=$1, city=$2, state=$3, zip_code=$4, country=$5 WHERE id=$6;

-- name: DeleteAddress :exec
DELETE FROM addresses WHERE id=$1;

-- name: SetDefaultAddress :exec
WITH reset AS (
    UPDATE addresses AS a SET is_default=false WHERE a.user_id=$1
)
UPDATE addresses AS a SET is_default=true WHERE a.id=$2;
-- name: GetPass :one
SELECT id, password, name, admin FROM users
 WHERE email = $1 LIMIT 1;

-- name: CreateUser :exec
INSERT INTO users (
  id, email, name, password
) VALUES (
  $1, $2, $3, $4
);

-- name: UpdateUser :exec
UPDATE users
 SET
  email = COALESCE(NULLIF($2, ''), email),
  name = COALESCE(NULLIF($3, ''), name)
WHERE email = $1;


-- name: PromoteAdmin :exec
UPDATE users SET admin = true
 WHERE email = $1;

-- name: DemoteAdmin :exec
UPDATE users SET admin = false
 WHERE email = $1;

-- name: CreateDestination :exec
INSERT INTO destination (
  id, name, description, attraction, pic_url
) VALUES (
  $1, $2, $3, $4, $5
);

-- name: ListDestinations :many
SELECT * FROM destination;

-- name: GetDestination :one
SELECT name, description, attraction FROM destination
 WHERE id = $1 LIMIT 1;

-- name: UpdateDestination :exec
UPDATE destination
 SET
  name = COALESCE(NULLIF($2, ''), name),
  description = COALESCE(NULLIF($3, ''), description),
  attraction = COALESCE(NULLIF($4, ''), attraction),
  pic_url = COALESCE(NULLIF($5, ''), pic_url)
WHERE id = $1;

-- name: DeleteDestination :exec
DELETE FROM destination
 WHERE id = $1;


-- name: CreateTrip :exec
INSERT INTO trip (
  id, name, start_date, end_date, destination_id
) VALUES (
  $1, $2, $3, $4, $5
);

-- name: ListTrips :many
SELECT * FROM trip;

-- name: GetTrip :one
SELECT name, start_date, end_date, destination_id FROM trip
 WHERE id = $1 LIMIT 1;

-- name: GetTripsByDestinationID :many
SELECT id, name, start_date, end_date FROM trip
 WHERE destination_id = $1;

-- name: UpdateTrip :exec
UPDATE trip
 SET
  name = COALESCE(NULLIF($2, ''), name),
  start_date = COALESCE(NULLIF($3::date, NULL), start_date),
  end_date = COALESCE(NULLIF($4::date, NULL), end_date),
  destination_id = COALESCE(NULLIF($5::uuid, NULL), destination_id)
WHERE id = $1;

-- name: DeleteTrip :exec
DELETE FROM trip
 WHERE id = $1;

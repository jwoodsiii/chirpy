-- name: CreateChirp :one
insert into chirps (id, created_at, updated_at, body, user_id)
values (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
returning *;

-- name: GetChirps :many
select * from chirps order by created_at asc;

-- name: GetChirp :one
select * from chirps where id=$1;

-- name: DeleteChirp :one
delete from chirps where id=$1 and user_id=$2
returning *;

-- name: GetChirpsByAuthor :many
select * from chirps where user_id=$1 order by created_at asc;

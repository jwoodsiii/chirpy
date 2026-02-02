-- name: CreateToken :one
insert into refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
values (
    $1,
    NOW(),
    NOW(),
    $2,
    NOW()+interval '60 day',
    null
)
returning *;

-- name: GetRefreshToken :one
select * from refresh_tokens where token=$1;

-- name: GetUserFromRefreshToken :one
select * from refresh_tokens
where token=$1
and expires_at > NOW()
and revoked_at is null;

-- name: RevokeToken :one
update refresh_tokens
set revoked_at=NOW(), updated_at=NOW()
where token=$1
returning *;

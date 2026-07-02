# Authentication

Email and password auth with opaque access tokens and rotating refresh
tokens, scoped to device sessions.

## Flow

1. `Register` or `Login` returns an access token, a refresh token, and
   the user.
2. The client sends `authorization: Bearer <access_token>` on each call.
3. When the access token expires, call `RefreshToken` to rotate it.

## Tokens

- **Access token**: random opaque string stored in Redis with a TTL. A
  request validates it with a single Redis lookup. No JWT parsing.
- **Refresh token**: only its SHA-256 hash is stored, in
  `auth.refresh_tokens`. Rotation issues a new token in the same family;
  reusing a revoked token revokes the whole family.

## Sessions

Every login opens a row in `auth.device_sessions`. Revoking a session
blocks the access token immediately. Users can list and revoke their
devices. The oldest session is evicted when the per-user device limit is
reached.

## Email verification and password reset

Verification and reset tokens are stored in Redis with a TTL. The email
worker delivers the message; the token maps back to the user on use.

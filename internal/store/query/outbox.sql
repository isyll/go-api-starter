-- name: InsertOutbox :one
INSERT INTO events.outbox (event_type, payload, dedupe_key)
VALUES ($1, $2, $3)
RETURNING id;

-- name: MarkOutboxProcessed :exec
UPDATE events.outbox SET processed_at = now() WHERE id = $1;

-- name: MarkOutboxFailed :exec
UPDATE events.outbox
SET retry_count = retry_count + 1, last_error = $2, last_attempted_at = now()
WHERE id = $1;

-- name: InsertOutboxDeadLetter :exec
INSERT INTO events.outbox_dead_letter (
  source_id, event_type, payload, failure_reason, last_error
) VALUES ($1, $2, $3, $4, $5);

-- name: PendingOutboxBatch :many
SELECT * FROM events.outbox
WHERE processed_at IS NULL
  AND retry_count < $1
  AND (
    last_attempted_at IS NULL
    OR (retry_count = 1 AND last_attempted_at < now() - interval '30 seconds')
    OR (retry_count = 2 AND last_attempted_at < now() - interval '2 minutes')
    OR (retry_count = 3 AND last_attempted_at < now() - interval '10 minutes')
    OR (retry_count = 4 AND last_attempted_at < now() - interval '30 minutes')
    OR (retry_count >= 5 AND last_attempted_at < now() - interval '1 hour')
  )
ORDER BY last_attempted_at NULLS FIRST, id ASC
LIMIT $2
FOR UPDATE SKIP LOCKED;

-- name: OutboxStats :one
SELECT
  COALESCE(SUM(CASE WHEN processed_at IS NULL AND retry_count < $1 THEN 1 ELSE 0 END), 0)::bigint AS pending,
  COALESCE(SUM(CASE WHEN processed_at IS NULL AND retry_count >= $1 THEN 1 ELSE 0 END), 0)::bigint AS exhausted,
  COALESCE(EXTRACT(EPOCH FROM (now() - MIN(CASE WHEN processed_at IS NULL AND retry_count < $1 THEN created_at END))), 0)::float8 AS oldest_pending_secs
FROM events.outbox;

-- name: CountOutboxDeadLetters :one
SELECT count(*) FROM events.outbox_dead_letter;

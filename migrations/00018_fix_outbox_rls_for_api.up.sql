-- The api role could INSERT outbox rows but nothing else, which broke
-- three legitimate paths under FORCE RLS:
--   - InsertOutbox ... RETURNING id (RETURNING needs a SELECT policy),
--   - MarkProcessed / MarkFailed after a direct dispatch,
--   - the drain_on_api kill switch (SELECT FOR UPDATE + UPDATE).
-- DELETE stays denied: the API can never purge queue history.

CREATE POLICY outbox_app_select ON events.outbox
FOR SELECT TO app_api USING (TRUE);

CREATE POLICY outbox_app_update ON events.outbox
FOR UPDATE TO app_api USING (TRUE) WITH CHECK (TRUE);

COMMENT ON POLICY outbox_app_select ON events.outbox IS
'app_api reads outbox rows: needed for INSERT..RETURNING and the drain_on_api kill switch.';

COMMENT ON POLICY outbox_app_update ON events.outbox IS
'app_api updates outbox rows: mark processed/failed after direct dispatch and when draining. DELETE remains denied.';

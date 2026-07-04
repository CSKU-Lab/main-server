CREATE OR REPLACE FUNCTION code_submissions_outbox_notify()
RETURNS trigger AS $$
BEGIN
  -- 'code_submissions_outbox_insert' is the channel name.
  -- Notify with the row id ONLY, not the full row. pg_notify has an 8000-byte
  -- payload cap; row_to_json(NEW) embeds the grade payload (all assembled source
  -- files), which can exceed it and make this trigger — and therefore the whole
  -- INSERT transaction — fail. The worker fetches the persisted row by id.
  PERFORM pg_notify(
    'code_submissions_outbox_insert',
    NEW.id::text
  );

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION code_submissions_outbox_notify()
RETURNS trigger AS $$
BEGIN
  -- 'code_submissions_outbox_insert' is the channel name
  -- payload is the JSON representation of the new row
  PERFORM pg_notify(
    'code_submissions_outbox_insert',
    row_to_json(NEW)::text
  );

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

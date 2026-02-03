CREATE OR REPLACE FUNCTION notify_row_update()
RETURNS trigger AS $$
BEGIN
  -- Notify a channel specific to this record's ID
  -- Example channel: 'user_update_42'
  PERFORM pg_notify(
    TG_TABLE_NAME || '_update_' || NEW.id,
    row_to_json(NEW)::text
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

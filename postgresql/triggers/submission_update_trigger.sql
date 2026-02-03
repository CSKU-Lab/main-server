CREATE TRIGGER submission_updated_trigger
AFTER UPDATE ON submissions
FOR EACH ROW
EXECUTE FUNCTION notify_row_update();

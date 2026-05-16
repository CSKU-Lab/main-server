CREATE OR REPLACE TRIGGER trigger_code_submissions_outbox_insert
AFTER INSERT ON code_submissions_outbox
FOR EACH ROW
EXECUTE FUNCTION code_submissions_outbox_notify();

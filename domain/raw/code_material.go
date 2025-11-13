package raw

type CodeMaterial struct {
	ID          string `db:"id"`
	Description string `db:"description"`
	TaskID      string `db:"task_id"`
}

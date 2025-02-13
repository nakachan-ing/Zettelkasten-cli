package internal

type Zettel struct {
	ID         string   `json:"id"`
	NoteID     string   `json:"note_id"`
	Title      string   `json:"title"`
	NoteType   string   `json:"note_type"`
	Tags       []string `json:"tags"`
	TaskStatus string   `json:"task_status"`
	Links      []string `json:"Links"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
	NotePath   string   `json:"note_path"`
	Archived   bool     `json:"archived"`
	Deleted    bool     `json:"deleted"`
}

package model

type Task struct {
	BaseModel
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

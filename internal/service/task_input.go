package service

type CreateTaskInput struct {
	Title       string
	Description string
}

type UpdateTaskInput struct {
	Title       *string
	Description *string
	Done        *bool
}

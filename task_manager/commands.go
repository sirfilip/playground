package task_manager

type CreateTaskCmd struct {
	Owner       Owner
	Title       TaskTitle
	Description TaskDescription
	TaskId      TaskID
	DueDate     DueDate
}

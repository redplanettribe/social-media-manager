package queue

type PostQueueStatus string

const (
	StatusPending    PostQueueStatus = "pending"
	StatusProcessing PostQueueStatus = "processing"
	StatusCompleted  PostQueueStatus = "completed"
	StatusFailed     PostQueueStatus = "failed"
)


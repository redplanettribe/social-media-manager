package post

// FIFO queue of postIDs
type Queue []string

// Add adds a postID to the queue
func (q *Queue) Add(postID string) {
	*q = append(*q, postID)
}

// Remove removes a postID from the queue
func (q *Queue) Remove(postID string) {
	for i, id := range *q {
		if id == postID {
			*q = append((*q)[:i], (*q)[i+1:]...)
			return
		}
	}
}

// RemoveAt removes a postID at a specific index from the queue
func (q *Queue) RemoveAt(index int) {
	*q = append((*q)[:index], (*q)[index+1:]...)
}

// InsertAt inserts a postID at a specific index in the queue. If the index is out of bounds, it will be added to the beginning or end of the queue
func (q *Queue) InsertAt(index int, postID string) {
	if index < 0 {
		index = 0
	}
	if index > len(*q) {
		index = len(*q)
	}
	*q = append((*q)[:index], append([]string{postID}, (*q)[index:]...)...)
}

// Get returns a postID at a specific index in the queue. If the index is out of bounds, it will return the first or last postID
func (q *Queue) Get(index int) string {
	if index < 0 {
		index = 0
	}
	if index > len(*q) {
		index = len(*q)
	}
	return (*q)[index]
}

// GetFirst returns the first postID in the queue
func (q *Queue) GetFirst() string {
	return q.Get(0)
}

// Shift removes the first postID in the queue and returns it
func (q *Queue) Shift() string {
	return q.Pop(0)
}

// Pop removes a postID at a specific index in the queue and returns it
func (q *Queue) Pop(index int) string {
	if index < 0 {
		index = 0
	}
	if index > len(*q) {
		index = len(*q)
	}
	postID := (*q)[index]
	*q = append((*q)[:index], (*q)[index+1:]...)
	return postID
}

// Move moves a postID from one index to another in the queue. If the from or to index is out of bounds, it will be moved to the beginning or end of the queue
func (q *Queue) Move(from, to int) {
	if from < 0 {
		from = 0
	}
	if from > len(*q) {
		from = len(*q)
	}
	if to < 0 {
		to = 0
	}
	if to > len(*q) {
		to = len(*q)
	}
	postID := (*q)[from]
	*q = append((*q)[:from], (*q)[from+1:]...)
	q.InsertAt(to, postID)
}

// Contains checks if a postID is in the queue
func (q *Queue) Contains(postID string) bool {
	for _, id := range *q {
		if id == postID {
			return true
		}
	}
	return false
}

// IsEmpty checks if the queue is empty
func (q *Queue) IsEmpty() bool {
	return len(*q) == 0
}

// Len returns the length of the queue
func (q *Queue) Len() int {
	return len(*q)
}

func (q *Queue) Arr() []string {
	return *q
}
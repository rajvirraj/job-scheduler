package scheduler

// JobQueue implements heap.Interface with a min heap using ScheduledJob.Next as the key.
type JobQueue []*ScheduledJob

func (h *JobQueue) Len() int           { return len(*h) }
func (h *JobQueue) Less(i, j int) bool { return (*h)[i].Next.Before((*h)[j].Next) }
func (h *JobQueue) Swap(i, j int)      { (*h)[i], (*h)[j] = (*h)[j], (*h)[i] }

func (h *JobQueue) Push(x any) {
	*h = append(*h, x.(*ScheduledJob))
}

func (h *JobQueue) Pop() any {
	current := *h
	n := len(current)
	minJob := current[n-1]
	*h = current[0 : n-1]
	return minJob
}

func (h *JobQueue) Peek() (*ScheduledJob, error) {
	if len(*h) == 0 {
		return nil, ErrJobQueueEmpty
	}
	return (*h)[0], nil
}

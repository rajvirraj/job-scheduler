package scheduler

import (
	"container/heap"
	"testing"
	"time"
)

func TestJobQueue_Push(t *testing.T) {
	jobQueue := &JobQueue{}
	heap.Init(jobQueue)

	job1 := &ScheduledJob{Id: "job1", Next: time.Now().Add(10 * time.Minute)}
	job2 := &ScheduledJob{Id: "job2", Next: time.Now().Add(5 * time.Minute)}
	job3 := &ScheduledJob{Id: "job3", Next: time.Now().Add(15 * time.Minute)}

	heap.Push(jobQueue, job1)
	heap.Push(jobQueue, job2)
	heap.Push(jobQueue, job3)

	if (*jobQueue)[0].Id != "job2" {
		t.Errorf("expected job2 to be at the root, got %s", (*jobQueue)[0].Id)
	}
}

func TestJobQueue_Pop(t *testing.T) {
	jobQueue := &JobQueue{}
	heap.Init(jobQueue)

	job1 := &ScheduledJob{Id: "job1", Next: time.Now().Add(10 * time.Minute)}
	job2 := &ScheduledJob{Id: "job2", Next: time.Now().Add(5 * time.Minute)}
	job3 := &ScheduledJob{Id: "job3", Next: time.Now().Add(15 * time.Minute)}

	heap.Push(jobQueue, job1)
	heap.Push(jobQueue, job2)
	heap.Push(jobQueue, job3)

	job := heap.Pop(jobQueue).(*ScheduledJob)
	if job.Id != "job2" {
		t.Errorf("expected job2 to be popped, got %s", job.Id)
	}
}

func TestJobQueue_Peek(t *testing.T) {
	jobQueue := &JobQueue{}
	heap.Init(jobQueue)

	job1 := &ScheduledJob{Id: "job1", Next: time.Now().Add(10 * time.Minute)}
	job2 := &ScheduledJob{Id: "job2", Next: time.Now().Add(5 * time.Minute)}
	job3 := &ScheduledJob{Id: "job3", Next: time.Now().Add(15 * time.Minute)}

	heap.Push(jobQueue, job1)
	heap.Push(jobQueue, job2)
	heap.Push(jobQueue, job3)

	job, peekErr := jobQueue.Peek()
	if peekErr != nil {
		t.Fatalf("unexpected error: %v", peekErr)
	}

	if job.Id != "job2" {
		t.Errorf("expected job2 to be peeked, got %s", job.Id)
	}
}

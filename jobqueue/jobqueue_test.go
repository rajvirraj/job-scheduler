package jobqueue

import (
	"container/heap"
	"testing"
	"time"

	"github.com/personal/assignment_2/runner"
)

func TestJobQueue_Push(t *testing.T) {
	jobQueue := &JobQueue{}
	heap.Init(jobQueue)

	job1 := &runner.ScheduledJob{Id: "job1", NextRunTime: time.Now().Add(10 * time.Minute)}
	job2 := &runner.ScheduledJob{Id: "job2", NextRunTime: time.Now().Add(5 * time.Minute)}
	job3 := &runner.ScheduledJob{Id: "job3", NextRunTime: time.Now().Add(15 * time.Minute)}

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

	job1 := &runner.ScheduledJob{Id: "job1", NextRunTime: time.Now().Add(10 * time.Minute)}
	job2 := &runner.ScheduledJob{Id: "job2", NextRunTime: time.Now().Add(5 * time.Minute)}
	job3 := &runner.ScheduledJob{Id: "job3", NextRunTime: time.Now().Add(15 * time.Minute)}

	heap.Push(jobQueue, job1)
	heap.Push(jobQueue, job2)
	heap.Push(jobQueue, job3)

	job := heap.Pop(jobQueue).(*runner.ScheduledJob)
	if job.Id != "job2" {
		t.Errorf("expected job2 to be popped, got %s", job.Id)
	}
}

func TestJobQueue_Peek(t *testing.T) {
	jobQueue := &JobQueue{}
	heap.Init(jobQueue)

	job1 := &runner.ScheduledJob{Id: "job1", NextRunTime: time.Now().Add(10 * time.Minute)}
	job2 := &runner.ScheduledJob{Id: "job2", NextRunTime: time.Now().Add(5 * time.Minute)}
	job3 := &runner.ScheduledJob{Id: "job3", NextRunTime: time.Now().Add(15 * time.Minute)}

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

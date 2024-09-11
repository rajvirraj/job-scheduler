package scheduler

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/personal/assignment_2/runner"
)

type TestRunner struct {
	wg           *sync.WaitGroup
	name         string
	sleepSeconds int
}

func (t *TestRunner) Run(timeout time.Duration) int32 {
	select {
	case <-time.After(timeout):
		return 2
	case <-time.After(time.Duration(t.sleepSeconds) * time.Second):
		t.wg.Done()
		return 0
	}
}

func TestScheduler_SingleJob(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(3)

	scheduler := NewScheduler()
	job := runner.NewScheduledJob("job1", &RepeatSchedule{repeatInterval: 5 * time.Second}, 10*time.Second, &TestRunner{wg: &wg, name: "TestJob1"})
	err := scheduler.AddJob(job)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	scheduler.Start()
	defer scheduler.Stop()
	timer := time.NewTimer(17 * time.Second)
	select {
	case <-timer.C:
		t.Fatal("expected job to run")
	case <-wait(&wg):
		fmt.Println("job1 done 7 times")
	}
	// to check that the job does not run more than 3 times in 17 seconds. If it does, it will panic
	<-timer.C
}

func TestScheduler_MultipleJobs(t *testing.T) {
	var wg1 sync.WaitGroup
	wg1.Add(3)

	scheduler := NewScheduler()
	job1 := runner.NewScheduledJob("job1", &RepeatSchedule{repeatInterval: 7 * time.Second}, 10*time.Second, &TestRunner{wg: &wg1, name: "TestJob1"})
	addJob1Err := scheduler.AddJob(job1)
	if addJob1Err != nil {
		t.Fatalf("expected no error, got %v", addJob1Err)
	}

	var wg2 sync.WaitGroup
	wg2.Add(4)
	job2 := runner.NewScheduledJob("job2", &RepeatSchedule{repeatInterval: 5 * time.Second}, 10*time.Second, &TestRunner{wg: &wg2, name: "TestJob2"})
	addJob2Err := scheduler.AddJob(job2)
	if addJob2Err != nil {
		t.Fatalf("expected no error, got %v", addJob2Err)
	}

	scheduler.Start()
	defer scheduler.Stop()

	var combinedWg sync.WaitGroup
	combinedWg.Add(2)
	timer := time.NewTimer(23 * time.Second)

	select {
	case <-timer.C:
		t.Fatal("expected job to run")
	case <-wait(&wg1):
		fmt.Println("job1 done 3 times")
		combinedWg.Done()
	}

	select {
	case <-timer.C:
		t.Fatal("expected job to run")
	case <-wait(&wg2):
		fmt.Println("job2 done 4 times")
		combinedWg.Done()
	}

	select {
	case <-timer.C:
		t.Fatal("expected number of jobs did not run in given time")
	case <-wait(&combinedWg):
		fmt.Println("both jobs done expected number of times")
	}
	// waiting for timer to complete. to check that the jobs do not run more than 3 and 4 times respectively in 23 seconds. If it does, panic will occur since wait.Done() will make counter negative
	<-timer.C
}

func TestScheduler_RemoveJob(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(3)

	scheduler := NewScheduler()
	job := runner.NewScheduledJob("job1", &RepeatSchedule{repeatInterval: 5 * time.Second}, 10*time.Second, &TestRunner{wg: &wg, name: "TestJob1"})
	err := scheduler.AddJob(job)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	scheduler.Start()
	defer scheduler.Stop()
	timer := time.NewTimer(17 * time.Second)
	select {
	case <-timer.C:
		t.Fatal("expected job to run")
	case <-wait(&wg):
		fmt.Println("job1 done 3 times")
	}
	// to check that the job does not run more than 3 times in 17 seconds. If it does, it will panic
	<-timer.C

	// delete job
	scheduler.RemoveJob("job1")
	// check if job is not running any further
	wg.Add(1)
	timer = time.NewTimer(30 * time.Second)
	select {
	case <-timer.C:
		fmt.Println("waited for 30s. As expected job did not run since it was removed")
	case <-wait(&wg):
		t.Fatal("did not expect job to run since it was removed")
	}
}

func TestScheduler_PauseAndResume(t *testing.T) {
	// write test case for pausing the scheduler
	var wg sync.WaitGroup
	wg.Add(3)

	scheduler := NewScheduler()
	job := runner.NewScheduledJob("job1", &RepeatSchedule{repeatInterval: 5 * time.Second}, 10*time.Second, &TestRunner{wg: &wg, name: "TestJob1"})
	err := scheduler.AddJob(job)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	scheduler.Start()
	defer scheduler.Stop()
	timer := time.NewTimer(17 * time.Second)
	select {
	case <-timer.C:
		t.Fatal("expected job to run")
	case <-wait(&wg):
		fmt.Println("job1 done 3 times")
	}
	// to check that the job does not run more than 3 times in 17 seconds. If it does, panic will occur since wait.Done() will make counter negative
	<-timer.C

	// pause scheduler
	scheduler.Pause()
	wg.Add(1)
	select {
	case <-time.After(10 * time.Second):
		fmt.Println("waited for 10s. As expected job did not run since scheduler is paused")
	case <-wait(&wg):
		t.Fatal("job should not have run")
	}

	//un pause scheduler.
	wg.Add(2) // add only 2 since we have added one before
	scheduler.Resume()
	timer = time.NewTimer(17 * time.Second)
	select {
	case <-timer.C:
		t.Fatal("expected job to run")
	case <-wait(&wg):
		fmt.Println("job1 done 3 times")
	}
	// to check that the job does not run more than 3 times in 17 seconds. If it does, it will panic
	<-timer.C
}

func TestScheduler_GetJobDetails(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(3)

	scheduler := NewScheduler()
	schedule := &RepeatSchedule{repeatInterval: 5 * time.Second}
	job := runner.NewScheduledJob("job1", schedule, 10*time.Second, &TestRunner{wg: &wg, name: "TestJob1"})
	err := scheduler.AddJob(job)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	scheduler.Start()
	defer scheduler.Stop()
	timer := time.NewTimer(17 * time.Second)
	select {
	case <-timer.C:
		t.Fatal("expected job to run")
	case <-wait(&wg):
		fmt.Println("job1 done 3 times")
	}
	// to check that the job does not run more than 3 times in 17 seconds. If it does, it will panic
	<-timer.C
	scheduler.Pause()
	jobDetails, getJobDetailsErr := scheduler.GetJobDetails("job1")
	if getJobDetailsErr != nil {
		t.Fatalf("expected no error, got %v", getJobDetailsErr)
	}
	fmt.Println(jobDetails)
	if job.Metrics.SuccessCount != 3 || job.Metrics.FailureCount != 0 || job.Metrics.ForceTerminatedCount != 0 || job.Metrics.FailedForceTerminationCount != 0 {
		t.Fatalf("expected value for JobMetrics. got %v", job.Metrics)
	}
}

func TestScheduler_MaxExecutionTimeTermination(t *testing.T) {
	var wg sync.WaitGroup

	scheduler := NewScheduler()
	schedule := &RepeatSchedule{repeatInterval: 5 * time.Second}
	job := runner.NewScheduledJob("job1", schedule, 2*time.Second, &TestRunner{wg: &wg, name: "TestJob1", sleepSeconds: 4})
	err := scheduler.AddJob(job)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	scheduler.Start()
	defer scheduler.Stop()
	timer := time.NewTimer(18 * time.Second)
	<-timer.C
	scheduler.Pause()
	jobDetails, getJobDetailsErr := scheduler.GetJobDetails("job1")
	if getJobDetailsErr != nil {
		t.Fatalf("expected no error, got %v", getJobDetailsErr)
	}
	if jobDetails.Metrics.SuccessCount != 0 || jobDetails.Metrics.FailureCount != 0 || jobDetails.Metrics.ForceTerminatedCount != 3 || jobDetails.Metrics.FailedForceTerminationCount != 0 {
		t.Fatalf("expected value for JobMetrics. got %v", job.Metrics)
	}
}

// Mock Schedule for testing
type RepeatSchedule struct {
	repeatInterval time.Duration
}

func (s *RepeatSchedule) Next(time.Time) time.Time {
	return time.Now().Add(s.repeatInterval)
}

func wait(wg *sync.WaitGroup) chan bool {
	ch := make(chan bool)
	go func() {
		wg.Wait()
		ch <- true
	}()
	return ch
}

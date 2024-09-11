package scheduler

import (
	"container/heap"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/personal/assignment_2/jobqueue"
	"github.com/personal/assignment_2/runner"
)

type schedulerState int32

const (
	schedulerInitialised schedulerState = iota
	schedulerRunning
	schedulerPaused
	schedulerStopped
)

type Scheduler struct {
	scheduledJobs          map[string]*runner.ScheduledJob
	deletedJobs            map[string]*runner.ScheduledJob
	schedulerState         schedulerState
	mutex                  *sync.RWMutex
	jobQueue               *jobqueue.JobQueue
	addJobChan             chan string
	stopSchedulerChan      chan struct{}
	pauseSchedulerChan     chan struct{}
	resumeSchedulerChannel chan struct{}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		scheduledJobs:          make(map[string]*runner.ScheduledJob),
		deletedJobs:            make(map[string]*runner.ScheduledJob),
		schedulerState:         schedulerInitialised,
		mutex:                  &sync.RWMutex{},
		addJobChan:             make(chan string),
		stopSchedulerChan:      make(chan struct{}),
		pauseSchedulerChan:     make(chan struct{}),
		resumeSchedulerChannel: make(chan struct{}),
	}
}

func (s *Scheduler) AddJob(job *runner.ScheduledJob) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, exists := s.deletedJobs[job.Id]; exists {
		return ErrJobIdPreviouslyDeleted
	}
	if _, exists := s.scheduledJobs[job.Id]; exists {
		return ErrJobAlreadyAdded
	}
	s.scheduledJobs[job.Id] = job
	if s.schedulerState == schedulerRunning {
		s.addJobChan <- job.Id
	}
	return nil
}

func (s *Scheduler) Start() {
	s.mutex.Lock()
	if s.schedulerState == schedulerRunning {
		return
	}
	s.schedulerState = schedulerRunning
	// build a min heap using the scheduled jobs
	s.jobQueue = buildJobQueue(s.scheduledJobs)
	s.mutex.Unlock()

	go func() {
		for {
			if s.schedulerState == schedulerStopped {
				break
			}
			curTime := time.Now()
			// determine the next entry to run.
			var timer *time.Timer
			nextJob, peekErr := s.jobQueue.Peek()
			if peekErr != nil {
				if errors.Is(peekErr, jobqueue.ErrJobQueueEmpty) {
					// no jobs in queue. Sleep for a long time. The scheduler will be woken up when a new job is added.
					// Todo : handle corner case on what happens when this wakes up. (i.e no job is added)
					timer = time.NewTimer(10000 * time.Hour)
				} else {
					fmt.Println("error in peek operation on job queue")
					continue
				}
			}
			if peekErr == nil {
				timer = time.NewTimer(nextJob.NextRunTime.Sub(curTime))
			}

			select {
			case curTime = <-timer.C:
				jobToRun := heap.Pop(s.jobQueue).(*runner.ScheduledJob)
				timer.Stop()
				if jobToRun.ScheduleStatus != runner.Active {
					continue
				}
				jobToRun.Run()
				jobToRun.NextRunTime = jobToRun.Schedule.Next(curTime)
				heap.Push(s.jobQueue, jobToRun)
			case jobId := <-s.addJobChan:
				timer.Stop()
				job := s.scheduledJobs[jobId]
				job.NextRunTime = job.Schedule.Next(curTime)
				heap.Push(s.jobQueue, job)
			case <-s.stopSchedulerChan:
				timer.Stop()
				fmt.Println("stopping scheduler")
			case <-s.pauseSchedulerChan:
				timer.Stop()
				// block this goroutine until resume is called
				fmt.Println("Scheduler : blocking on resuming scheduler")
				select {
				case <-s.resumeSchedulerChannel:
					fmt.Println("Scheduler : resuming scheduler")
					s.jobQueue = buildJobQueue(s.scheduledJobs)
				case <-s.stopSchedulerChan:
					fmt.Println("Scheduler : stopping scheduler")
				}
			}
		}
	}()
}

func buildJobQueue(scheduledJobs map[string]*runner.ScheduledJob) *jobqueue.JobQueue {
	curTime := time.Now()
	jobQueue := &jobqueue.JobQueue{}
	heap.Init(jobQueue)
	for _, job := range scheduledJobs {
		job.NextRunTime = job.Schedule.Next(curTime)
		heap.Push(jobQueue, job)
	}
	return jobQueue
}

func (s *Scheduler) Pause() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.schedulerState == schedulerPaused {
		return nil
	}
	if s.schedulerState != schedulerRunning {
		return ErrSchedulerNotRunning(s.schedulerState)
	}
	s.schedulerState = schedulerPaused
	s.pauseSchedulerChan <- struct{}{}
	return nil
}

func (s *Scheduler) Resume() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.schedulerState != schedulerPaused {
		return ErrSchedulerNotRunning(s.schedulerState)
	}
	s.schedulerState = schedulerRunning
	s.resumeSchedulerChannel <- struct{}{}
	return nil
}

func (s *Scheduler) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.schedulerState == schedulerStopped {
		return nil
	}
	s.schedulerState = schedulerStopped
	s.stopSchedulerChan <- struct{}{}
	return nil
}

func (s *Scheduler) GetJobDetails(jobId string) (JobDetails, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if job, exists := s.scheduledJobs[jobId]; exists {
		return getJobDetails(job), nil
	}
	return JobDetails{}, ErrJobNotFound
}

func (s *Scheduler) RemoveJob(jobId string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if jobSchedule, exists := s.scheduledJobs[jobId]; exists {
		s.deletedJobs[jobId] = jobSchedule
		jobSchedule.DeleteSchedule()
	}
}

type JobDetails struct {
	Id               string
	Schedule         runner.Schedule
	MaxExecutionTime time.Duration
	Next             time.Time
	Metrics          *runner.JobMetrics
	JobStatus        runner.JobStatus
}

func getJobDetails(job *runner.ScheduledJob) JobDetails {
	return JobDetails{
		Id:               job.Id,
		Schedule:         job.Schedule,
		MaxExecutionTime: job.MaxExecutionTime,
		Next:             job.NextRunTime,
		Metrics:          job.Metrics,
		JobStatus:        job.JobStatus,
	}
}

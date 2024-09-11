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
	// check if the job has been deleted before else return error
	if _, exists := s.deletedJobs[job.Id]; exists {
		return ErrJobIdPreviouslyDeleted
	}
	// add to scheduled jobs if not already present
	if _, exists := s.scheduledJobs[job.Id]; exists {
		return ErrJobAlreadyAdded
	}
	s.scheduledJobs[job.Id] = job
	// if the scheduler is running, send the job id to the addJobChan for the scheduler to add it to the job queue
	if s.schedulerState == schedulerRunning {
		s.addJobChan <- job.Id
	}
	return nil
}

func (s *Scheduler) Start() {
	s.mutex.Lock()
	// return if scheduler is already running
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
				return
			}
			curTime := time.Now()
			// determine the next entry to run.
			var timer *time.Timer
			nextJob, peekErr := s.jobQueue.Peek()
			if peekErr != nil {
				if errors.Is(peekErr, jobqueue.ErrJobQueueEmpty) {
					// no jobs in queue. Sleep for a long time. The scheduler will be woken up when a new job is added or after below timeout
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
				if s.jobQueue.Len() == 0 { // case when scheduler wakes up but no jobs in queue
					break
				}
				jobToRun := heap.Pop(s.jobQueue).(*runner.ScheduledJob)
				timer.Stop()
				if jobToRun.ScheduleStatus != runner.Active {
					continue
				}
				// run the job async and calculate the next run time for the job and push it to the job queue
				jobToRun.Run()
				jobToRun.NextRunTime = jobToRun.Schedule.Next(curTime)
				heap.Push(s.jobQueue, jobToRun)
			case jobId := <-s.addJobChan:
				timer.Stop()
				job := s.scheduledJobs[jobId]
				// calculate the next run time for the job and push it to the job queue
				job.NextRunTime = job.Schedule.Next(curTime)
				heap.Push(s.jobQueue, job)
			case <-s.stopSchedulerChan:
				timer.Stop()
				fmt.Println("stopping scheduler")
				return
			case <-s.pauseSchedulerChan:
				timer.Stop()
				// block this goroutine until resume is called
				fmt.Println("Scheduler : blocking on resuming scheduler")
				select {
				case <-s.resumeSchedulerChannel:
					fmt.Println("Scheduler : resuming scheduler")
					// rebuild job queue using current time
					s.jobQueue = buildJobQueue(s.scheduledJobs)
				case <-s.stopSchedulerChan:
					fmt.Println("Scheduler : stopping scheduler")
					return
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
	NextRunTime      time.Time
	Metrics          *runner.JobMetrics
	JobStatus        runner.JobStatus
}

func getJobDetails(job *runner.ScheduledJob) JobDetails {
	return JobDetails{
		Id:               job.Id,
		Schedule:         job.Schedule,
		MaxExecutionTime: job.MaxExecutionTime,
		NextRunTime:      job.NextRunTime,
		Metrics:          job.Metrics,
		JobStatus:        job.JobStatus,
	}
}

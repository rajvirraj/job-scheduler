package runner

import (
	"strconv"
	"sync"
	"time"
)

type Schedule interface {
	Next(t time.Time) time.Time
}

type ScheduledJob struct {
	Id               string
	Runnable         Runnable
	Schedule         Schedule
	MaxExecutionTime time.Duration
	NextRunTime      time.Time
	Metrics          *JobMetrics
	JobStatus        JobStatus
	mutex            *sync.RWMutex
	ScheduleStatus   ScheduleStatus
}

type JobStatus int32

const (
	// job is currently running
	Running JobStatus = iota
	// job is waiting according to its schedule
	Waiting
)

type ScheduleStatus int32

const (
	// schedule is active
	Active ScheduleStatus = iota
	// schedule is deleted
	Deleted
)

func NewScheduledJob(id string, schedule Schedule, maxExecutionTime time.Duration, job Runnable) *ScheduledJob {
	return &ScheduledJob{
		Id:               id,
		Runnable:         job,
		Schedule:         schedule,
		MaxExecutionTime: maxExecutionTime,
		Metrics: &JobMetrics{
			SuccessCount:               0,
			FailureCount:               0,
			ForceTerminatedCount:       0,
			AvgExecutionTimeForSuccess: 0.0,
			AvgExecutionTimeForFailure: 0.0,
		},
		JobStatus:      Waiting,
		mutex:          &sync.RWMutex{},
		ScheduleStatus: Active,
	}
}

type JobMetrics struct {
	SuccessCount                int
	ForceTerminatedCount        int
	FailureCount                int
	FailedForceTerminationCount int
	AvgExecutionTimeForSuccess  float32
	AvgExecutionTimeForFailure  float32
}

// this schedule only takes secinds as input
type CronSchedule struct {
	// repeat interval in seconds for job
	seconds int
}

func NewCronSchedule(expression string) *CronSchedule {
	seconds, convErr := strconv.Atoi(expression)
	if convErr != nil || seconds <= 0 {
		// Todo : handle error gracefully
		panic(convErr)
	}
	return &CronSchedule{seconds: seconds}
}

func (c *CronSchedule) Next(t time.Time) time.Time {
	return t.Add(time.Duration(c.seconds) * time.Second)
}

func (scheduledJob *ScheduledJob) Run() {
	go func() {
		now := time.Now()
		// check if job is already running, then return as change job status to running
		scheduledJob.mutex.Lock()
		if scheduledJob.JobStatus == Running {
			scheduledJob.mutex.Unlock()
			return
		}
		scheduledJob.JobStatus = Running
		scheduledJob.mutex.Unlock()

		// execute the job
		jobStatusCode := scheduledJob.Runnable.Run(scheduledJob.MaxExecutionTime)

		// acquire lock and update job_status and metrics
		scheduledJob.mutex.Lock()
		scheduledJob.JobStatus = Waiting
		scheduledJob.Metrics.updateMetrics(jobStatusCode, time.Since(now))
		scheduledJob.mutex.Unlock()
	}()
}

func (j *JobMetrics) updateMetrics(jobStatusCode int32, jobDuration time.Duration) {
	if jobStatusCode == 0 {
		j.SuccessCount++
		j.AvgExecutionTimeForSuccess = calculateNewAvgExecutionTime(j.AvgExecutionTimeForSuccess, j.SuccessCount, float32(jobDuration))
	} else if jobStatusCode == 1 {
		j.FailureCount++
		j.AvgExecutionTimeForSuccess = calculateNewAvgExecutionTime(j.AvgExecutionTimeForFailure, j.FailureCount, float32(jobDuration))
	} else if jobStatusCode == 2 {
		j.ForceTerminatedCount++
	} else {
		j.FailedForceTerminationCount++
	}
}

func calculateNewAvgExecutionTime(curAvgExecutionTime float32, curRunCount int, newExecutionTime float32) float32 {
	return ((curAvgExecutionTime * float32(curRunCount)) + newExecutionTime) / float32(curRunCount+1)
}

func (scheduledJob *ScheduledJob) GetScheduleStatus() ScheduleStatus {
	scheduledJob.mutex.RLock()
	defer scheduledJob.mutex.RUnlock()
	return scheduledJob.ScheduleStatus
}

func (scheduledJob *ScheduledJob) DeleteSchedule() {
	scheduledJob.mutex.Lock()
	defer scheduledJob.mutex.Unlock()
	scheduledJob.ScheduleStatus = Deleted
}

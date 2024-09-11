# Scheduler

## Overview

The Scheduler project is a Go-based job scheduling system that allows you to schedule and run jobs based on a schedule.

## Features

- **Job Scheduling**: Schedule jobs to run based on a schedule. You can create custom schedules by implementing the `Schedule` interface.
- **Job Execution**: Execute jobs based on a schedule with a specified timeout.
- **Job Management**: Add and remove jobs from the scheduler.
- **Scheduler Management**: Start, stop, pause and resume the scheduler.

## Usage


### Creating a Scheduler

```go
jobScheduler := scheduler.NewScheduler()
```

### Creating a Runnable Job

To create a runnable job, implement the `Runnable` interface and define the `Run` method which returns an exit code. This method should exit after ```timeout``` 
Below are the possible exit codes:
- 0: Script completed without any errors
- 1: Script returned an error
- 2: Job killed successfully after reaching timeout
- 3: Failed to kill the job after timeout

```go
type TestJob struct {
    name string
}

func (t *TestJob) Run(timeout time.Duration) int32 {
    fmt.Printf("Running job: %s\n", t.name)
    return 0
}
```
### Creating a Schedule

Implement the `Schedule` interface to create custom schedules:

```go
type SimpleSchedule struct{}

func (s *SimpleSchedule) Next(t time.Time) time.Time {
    return t.Add(1 * time.Hour)
}
```

or use existing `CronSchedule`. Below is a schedule to run every 5 seconds. 
Todo : extend this to accept cron expression
```go
schedule = runner.NewCronSchedule("5")
```

### Scheduling a Job and adding to Scheduler

Now that we have created a ```Runnable``` and a ```Schedule```, we can now create a `ScheduledJob` as below which can then be given to scheduler for running

```go
timeout := 10 * time.Second
runner := &TestJob{name: "job-1"}
scheduledJob1 := runner.NewScheduledJob("job-id-1", schedule, timeout, runner)
jobScheduler.AddJob(scheduledJob1)
```

### Starting the Scheduler
Once the scheduler has started, it will run the jobs based on the schedule. We can also add jobs after the scheduler has started.
```go
jobScheduler.Start()
```

### Stopping the Scheduler
Jobs will not be executed after the scheduler is stopped.
```go
jobScheduler.Stop()
```

### Pausing the Scheduler
Jobs will not be executed after the scheduler is paused. Scheduler can be resumed by calling `Resume` method.
```go
jobScheduler.Pause()
```

### Resuming the Scheduler

```go
jobScheduler.Resume()
```

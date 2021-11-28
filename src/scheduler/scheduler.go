package scheduler

import (
	"fmt"

	"github.com/robfig/cron/v3"
	"github.com/rymdo/kube-cron-rollout-restart/v2/src/types"
)

type Entry struct {
	Id  cron.EntryID
	Job types.Job
}

type OnExecute func(job types.Job)

type Scheduler struct {
	cron      *cron.Cron
	entries   []Entry
	onExecute OnExecute
}

func New(onExecute OnExecute) Scheduler {
	scheduler := Scheduler{
		cron:      cron.New(),
		onExecute: onExecute,
	}
	scheduler.cron.Start()
	return scheduler
}

func (s *Scheduler) AddJob(job types.Job) {
	if s.JobExists(job) {
		return
	}
	fmt.Printf("Scheduler: Adding job %+v\n", job)
	id, err := s.cron.AddFunc(job.Schedule, func() {
		fmt.Printf("Scheduler: Executing job %+v\n", job)
		s.onExecute(job)
	})
	if err != nil {
		fmt.Printf("Scheduler: Failed to add cron for job %+v\n", job)
		fmt.Println(err.Error())
		return
	}
	entry := Entry{
		Id:  id,
		Job: job,
	}
	s.entries = append(s.entries, entry)
}

func (s *Scheduler) RemoveJob(job types.Job) {
	fmt.Printf("Scheduler: Removing job %+v\n", job)
	for i, entry := range s.entries {
		if jobsEqual(entry.Job, job) {
			s.cron.Remove(entry.Id)
			s.entries = append(s.entries[:i], s.entries[i+1:]...)
		}
	}
}

func (s *Scheduler) JobExists(job types.Job) bool {
	for _, entry := range s.entries {
		if jobsEqual(entry.Job, job) {
			return true
		}
	}
	return false
}

func (s *Scheduler) SyncJobs(jobs []types.Job) {
	// Add missing jobs
	for _, job := range jobs {
		if !s.JobExists(job) {
			s.AddJob(job)
		}
	}

	// Clear missing jobs
	toRemove := []types.Job{}
	for _, entry := range s.entries {
		if !containsJob(jobs, entry.Job) {
			toRemove = append(toRemove, entry.Job)
		}
	}
	for _, job := range toRemove {
		s.RemoveJob(job)
	}
}

func jobsEqual(a types.Job, b types.Job) bool {
	if a.Namespace != b.Namespace {
		return false
	}
	if a.Workload != b.Workload {
		return false
	}
	if a.Schedule != b.Schedule {
		return false
	}
	return true
}

func containsJob(jobs []types.Job, targetJob types.Job) bool {
	for _, job := range jobs {
		if jobsEqual(job, targetJob) {
			return true
		}
	}
	return false
}

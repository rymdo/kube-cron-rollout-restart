package main

import (
	"time"

	"github.com/rymdo/kube-cron-rollout-restart/v2/src/kubernetes"
	"github.com/rymdo/kube-cron-rollout-restart/v2/src/scheduler"
	"github.com/rymdo/kube-cron-rollout-restart/v2/src/types"
)

func main() {
	k := kubernetes.New()
	s := scheduler.New(func(job types.Job) {
		k.RolloutRestart(job)
	})
	for true {
		jobs := k.GetJobs()
		s.SyncJobs(jobs)
		time.Sleep(time.Minute)
	}
}

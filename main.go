package main

import (
	"os"
	"time"

	"github.com/rymdo/kube-cron-rollout-restart/v2/src/kubernetes"
	"github.com/rymdo/kube-cron-rollout-restart/v2/src/scheduler"
	"github.com/rymdo/kube-cron-rollout-restart/v2/src/types"
)

func main() {
	useKubeConfig := false
	if os.Getenv("USE_KUBECONFIG") == "true" {
		useKubeConfig = true
	}

	k := kubernetes.New(useKubeConfig)
	s := scheduler.New(func(job types.Job) {
		k.RolloutRestart(job)
	})
	for true {
		jobs := k.GetJobs()
		s.SyncJobs(jobs)
		time.Sleep(time.Minute)
	}
}

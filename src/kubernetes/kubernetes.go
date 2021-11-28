package kubernetes

import (
	"fmt"

	"github.com/rymdo/kube-cron-rollout-restart/v2/src/types"
)

// cron.rollout.restart/schedule=* * * * *
// cron.rollout.restart/schedule=0 * * * *

type Kubernetes struct{}

func New() Kubernetes {
	return Kubernetes{}
}

func (k *Kubernetes) GetJobs() []types.Job {
	jobs := []types.Job{}
	jobs = append(jobs, types.Job{Namespace: "0", Workload: "1", Schedule: "* * * * *"})
	jobs = append(jobs, types.Job{Namespace: "0", Workload: "2", Schedule: "* * * * *"})
	jobs = append(jobs, types.Job{Namespace: "0", Workload: "3", Schedule: "* * * * *"})
	return jobs
}

func (k *Kubernetes) RolloutRestart(job types.Job) {
	fmt.Printf("Kubernetes: Restarting %+v\n", job)
}

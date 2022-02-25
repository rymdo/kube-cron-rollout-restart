package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rymdo/kube-cron-rollout-restart/v2/src/alertmanager"
	"github.com/rymdo/kube-cron-rollout-restart/v2/src/kubernetes"
	"github.com/rymdo/kube-cron-rollout-restart/v2/src/scheduler"
	"github.com/rymdo/kube-cron-rollout-restart/v2/src/types"

	"gopkg.in/alecthomas/kingpin.v2"
)

func getHomeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}
	return dir
}

func getDefaultKubeconfigDir() string {
	return fmt.Sprintf("%s/.kube/config", getHomeDir())
}

var (
	kubeconfigUse   = kingpin.Flag("kubeconfig", "use kubeconfig").Bool()
	kubeconfigPath  = kingpin.Flag("kubeconfig-path", "path to kubeconfig").Default(getDefaultKubeconfigDir()).String()
	alertmanagerUrl = kingpin.Flag("alertmanager-url", "url to alertmanger").Default("http://alertmanager:80").String()
)

func main() {
	fmt.Println("starting")

	kingpin.Parse()

	a := alertmanager.New(alertmanager.Config{Url: *alertmanagerUrl})
	k := kubernetes.New(*kubeconfigUse, *kubeconfigPath)
	s := scheduler.New(func(job types.Job) {
		if job.AlertmangerSilence != nil {
			err := a.CreateSilence(*job.AlertmangerSilence)
			if err != nil {
				fmt.Printf("scheduler/alertmanager: error %s\n", err.Error())
			}
		}
		k.RolloutRestart(job)
	})
	for true {
		jobs := k.GetJobs()
		s.SyncJobs(jobs)
		time.Sleep(time.Minute)
	}
}

package kubernetes

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rymdo/kube-cron-rollout-restart/v2/src/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const AnnotationScheduleKey = "cron.rollout.restart/schedule"

const AnnotationAlertmanagerSilenceEnabledKey = "cron.rollout.restart/alertmanager-silence-enabled"   // "true" or "false", default "false"
const AnnotationAlertmanagerSilenceDurationKey = "cron.rollout.restart/alertmanager-silence-duration" // duration in minutes
const AnnotationAlertmanagerSilenceLabelsKey = "cron.rollout.restart/alertmanager-silence-labels"     // comma separated silence matching labels, eg. key1=value1,key2=value2
const AnnotationAlertmanagerSilenceCommentKey = "cron.rollout.restart/alertmanager-silence-comment"   // comment

type Kubernetes struct {
	client *kubernetes.Clientset
}

func New(kubeconfigUse bool, kubeconfigPath string) Kubernetes {
	var config *rest.Config
	if kubeconfigUse {
		fmt.Printf("kubernetes: using kubeconfig mode - '%s'\n", kubeconfigPath)
		config = configKubeconfig(kubeconfigPath)
	} else {
		fmt.Println("kubernetes: using in-cluster mode")
		config = configInCluster()
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return Kubernetes{
		client: clientset,
	}
}

func configInCluster() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	return config
}

func configKubeconfig(kubeconfigPath string) *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err.Error())
	}
	return config
}

func parseAlertmanagerSilence(annotations map[string]string) *types.AlertmanagerSilence {
	// Check if enabled
	enabled := false
	duration := 0
	labels := ""
	comment := ""
	for key, value := range annotations {
		if key == AnnotationAlertmanagerSilenceEnabledKey {
			enabled = value == "true"
		}
		if key == AnnotationAlertmanagerSilenceDurationKey {
			if i, err := strconv.Atoi(value); err == nil {
				duration = i
			}
		}
		if key == AnnotationAlertmanagerSilenceLabelsKey {
			labels = value
		}
		if key == AnnotationAlertmanagerSilenceCommentKey {
			comment = value
		}
	}
	if !enabled {
		return nil
	}
	return &types.AlertmanagerSilence{
		Duration: duration,
		Labels:   labels,
		Comment:  comment,
	}
}

func (k *Kubernetes) GetJobs() []types.Job {
	jobs := []types.Job{}

	deployments, err := k.client.AppsV1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, deployment := range deployments.Items {
		for key, value := range deployment.Annotations {
			if key == AnnotationScheduleKey {
				jobs = append(jobs, types.Job{
					Namespace:           deployment.Namespace,
					Type:                "deployment",
					Workload:            deployment.Name,
					Schedule:            value,
					AlertmanagerSilence: parseAlertmanagerSilence(deployment.Annotations),
				})
			}
		}
	}

	statefulsets, err := k.client.AppsV1().StatefulSets("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, statefulset := range statefulsets.Items {
		for key, value := range statefulset.Annotations {
			if key == AnnotationScheduleKey {
				jobs = append(jobs, types.Job{
					Namespace:           statefulset.Namespace,
					Type:                "statefulset",
					Workload:            statefulset.Name,
					Schedule:            value,
					AlertmanagerSilence: parseAlertmanagerSilence(statefulset.Annotations),
				})
			}
		}
	}

	daemonsets, err := k.client.AppsV1().DaemonSets("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, daemonset := range daemonsets.Items {
		for key, value := range daemonset.Annotations {
			if key == AnnotationScheduleKey {
				jobs = append(jobs, types.Job{
					Namespace:           daemonset.Namespace,
					Type:                "daemonset",
					Workload:            daemonset.Name,
					Schedule:            value,
					AlertmanagerSilence: parseAlertmanagerSilence(daemonset.Annotations),
				})
			}
		}
	}
	return jobs
}

func (k *Kubernetes) RolloutRestart(job types.Job) {
	fmt.Printf("Kubernetes: Restarting %+v\n", job)
	data := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, time.Now().Format(time.RFC3339))
	if job.Type == "deployment" {
		_, err := k.client.AppsV1().Deployments(job.Namespace).Patch(context.TODO(), job.Workload, k8sTypes.StrategicMergePatchType, []byte(data), metav1.PatchOptions{FieldManager: "kubectl-rollout"})
		if err != nil {
			fmt.Printf("Kubernetes: Restart error %s\n", err.Error())
		}
	}
	if job.Type == "statefulset" {
		_, err := k.client.AppsV1().StatefulSets(job.Namespace).Patch(context.TODO(), job.Workload, k8sTypes.StrategicMergePatchType, []byte(data), metav1.PatchOptions{FieldManager: "kubectl-rollout"})
		if err != nil {
			fmt.Printf("Kubernetes: Restart error %s\n", err.Error())
		}
	}
	if job.Type == "daemonset" {
		_, err := k.client.AppsV1().DaemonSets(job.Namespace).Patch(context.TODO(), job.Workload, k8sTypes.StrategicMergePatchType, []byte(data), metav1.PatchOptions{FieldManager: "kubectl-rollout"})
		if err != nil {
			fmt.Printf("Kubernetes: Restart error %s\n", err.Error())
		}
	}
}

package kubernetes

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/rymdo/kube-cron-rollout-restart/v2/src/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const AnnotationScheduleKey = "cron.rollout.restart/schedule"

type Kubernetes struct {
	client *kubernetes.Clientset
}

func New(useKubeConfig bool) Kubernetes {
	var config *rest.Config
	if useKubeConfig {
		config = configKubeconfig()
	} else {
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

func configKubeconfig() *rest.Config {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config
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
					Namespace: deployment.Namespace,
					Type:      "deployment",
					Workload:  deployment.Name,
					Schedule:  value,
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
					Namespace: statefulset.Namespace,
					Type:      "statefulset",
					Workload:  statefulset.Name,
					Schedule:  value,
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
					Namespace: daemonset.Namespace,
					Type:      "daemonset",
					Workload:  daemonset.Name,
					Schedule:  value,
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

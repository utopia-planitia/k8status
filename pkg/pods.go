package k8status

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func PrintPodStatus(ctx context.Context, restconfig *rest.Config, clientset *kubernetes.Clientset, verbose bool) (int, error) {
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	healthy := 0
	total := 0
	for _, item := range pods.Items {
		if item.Status.Phase == v1.PodSucceeded || item.Status.Phase == v1.PodFailed {
			continue
		}

		total++

		containerReady := 0
		for _, containerStatus := range item.Status.ContainerStatuses {
			if containerStatus.Ready {
				containerReady++
			}
		}

		if len(item.Spec.Containers) == containerReady {
			healthy++
		}
	}

	fmt.Printf("%d of %d pods are running.\n", healthy, total)

	return 0, err
}

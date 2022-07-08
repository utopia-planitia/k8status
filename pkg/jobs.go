package k8status

import (
	"context"
	"fmt"

	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func PrintJobStatus(ctx context.Context, restconfig *rest.Config, clientset *kubernetes.Clientset, verbose bool) (int, error) {
	jobs, err := clientset.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	healthy := 0
	for _, item := range jobs.Items {
		if !isHealthy(item) {
			continue
		}

		healthy++
	}

	fmt.Printf("%d of %d jobs are completed.\n", healthy, len(jobs.Items))

	if verbose {
		for _, item := range jobs.Items {
			if isHealthy(item) {
				continue
			}

			fmt.Printf("Job %s in namespace %s failed.\n", item.Namespace, item.Name)
		}
	}

	return 0, nil
}

func isHealthy(item v1.Job) bool {
	if item.Status.Active > 0 {
		return true
	}

	if *item.Spec.Completions == item.Status.Succeeded {
		return true
	}

	return false
}

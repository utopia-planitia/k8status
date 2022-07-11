package k8status

import (
	"context"
	"fmt"
	"io"
	"strings"

	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintJobStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose bool) (int, error) {
	jobs, err := client.clientset.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
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

	fmt.Fprintf(header, "%d of %d jobs are completed.\n", healthy, len(jobs.Items))

	if verbose {
		for _, item := range jobs.Items {
			if isHealthy(item) {
				continue
			}

			fmt.Fprintf(details, "Job %s in namespace %s failed.\n", item.Namespace, item.Name)
		}
	}

	for _, item := range jobs.Items {
		if strings.Contains(item.ObjectMeta.Namespace, "ci") || strings.Contains(item.ObjectMeta.Namespace, "lab") {
			continue
		}

		if isHealthy(item) {
			continue
		}

		return 49, nil
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

package k8status

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func PrintJobStatus(ctx context.Context, clientset *kubernetes.Clientset, verbose bool) error {
	jobs, err := clientset.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	healthy := 0
	for _, item := range jobs.Items {
		if item.Status.Failed > 0 {
			continue
		}

		healthy++
	}

	fmt.Printf("%d of %d jobs are completed.\n", healthy, len(jobs.Items))
	return nil
}

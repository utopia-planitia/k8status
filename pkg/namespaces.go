package k8status

import (
	"context"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintNamespaceStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose bool) (int, error) {
	namespaces, err := client.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	healthy := 0
	for _, item := range namespaces.Items {
		if item.Status.Phase != v1.NamespaceActive {
			continue
		}

		healthy++
	}

	fmt.Fprintf(header, "%d of %d namespaces are active.\n", healthy, len(namespaces.Items))

	if healthy != len(namespaces.Items) {
		for _, item := range namespaces.Items {
			if item.Status.Phase != v1.NamespaceActive {
				fmt.Fprintf(details, "Namespace %s has status %s\n", item.Name, item.Status.Phase)
			}
		}

		return 43, nil
	}

	return 0, nil
}

package k8status

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func PrintNamespaceStatus(ctx context.Context, restconfig *rest.Config, clientset *kubernetes.Clientset, verbose bool) (int, error) {
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
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

	fmt.Printf("%d of %d namespaces are active.\n", healthy, len(namespaces.Items))

	if healthy != len(namespaces.Items) {
		return 43, nil
	}

	return 0, nil
}

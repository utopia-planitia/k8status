package k8status

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func PrintNamespaceStatus(ctx context.Context, clientset *kubernetes.Clientset, verbose bool) error {
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	healthy := 0
	for _, item := range namespaces.Items {
		if item.Status.Phase != v1.NamespaceActive {
			continue
		}

		healthy++
	}

	fmt.Printf("%d of %d namespaces are active.\n", healthy, len(namespaces.Items))

	return nil
}

package k8status

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func PrintCassandraStatus(ctx context.Context, restconfig *rest.Config, clientset *kubernetes.Clientset, verbose bool) (int, error) {

	exists, err := namespaceExists(ctx, clientset, "cassandra")
	if err != nil {
		return 0, err
	}

	if !exists {
		if verbose {
			fmt.Printf("Cassandra was not found.\n")
		}

		return 0, nil
	}

	// TODO check cassandra

	return 0, nil
}

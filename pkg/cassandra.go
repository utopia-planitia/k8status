package k8status

import (
	"context"
	"fmt"
	"k8s.io/client-go/kubernetes"
)

func PrintCassandraStatus(ctx context.Context, clientset *kubernetes.Clientset, verbose bool) error {

	exists, err := namespaceExists(ctx, clientset, "cassandra")
	if err != nil {
		return err
	}

	if !exists {
		if verbose {
			fmt.Printf("Cassandra was not found.\n")
		}
		return nil
	}

	// TODO check cassandra

	return nil
}

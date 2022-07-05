package k8status

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func PrintCassandraStatus(ctx context.Context, clientset *kubernetes.Clientset, verbose bool) error {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, "cassandra", metav1.GetOptions{})
	if err.Error() == "namespaces \"cassandra\" not found" {
		if verbose {
			fmt.Printf("Cassandra was not found.\n")
		}
		return nil
	}
	if err != nil {
		return err
	}

	// TODO check cassandra

	return nil
}

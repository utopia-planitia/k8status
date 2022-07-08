package k8status

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	cassandraNamespace = "cassandra"
)

func PrintCassandraStatus(ctx context.Context, restconfig *rest.Config, clientset *kubernetes.Clientset, verbose bool) (int, error) {

	exists, err := namespaceExists(ctx, clientset, cassandraNamespace)
	if err != nil {
		return 0, err
	}

	if !exists {
		if verbose {
			fmt.Printf("Cassandra was not found.\n")
		}

		return 0, nil
	}

	secret, err := clientset.CoreV1().Secrets(cassandraNamespace).Get(ctx, "k8ssandra-superuser", metav1.GetOptions{})
	username := secret.Data["username"]
	password := secret.Data["password"]

	output := &bytes.Buffer{}
	err = exec(
		clientset,
		restconfig,
		cassandraNamespace,
		"k8ssandra-dc1-default-sts-0",
		"cassandra",
		fmt.Sprintf("nodetool -u %s -pw %s --host ::FFFF:127.0.0.1 status | grep --extended-regexp '^[UD][NLJM]\\s+'", username, password),
		output,
	)

	readyNodes := strings.Count(output.String(), "UN")
	totalNodes := strings.Count(output.String(), "\n")

	fmt.Printf("%d of %d cassandra nodes are ready.\n", readyNodes, totalNodes)

	if verbose {
		fmt.Printf("%s", output.String())
	}

	if readyNodes != totalNodes {
		return 46, nil
	}

	return 0, nil
}

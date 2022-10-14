package k8status

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	cassandraNamespace = "cassandra"
)

func PrintCassandraStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose, colored bool) (int, error) {
	exists, err := namespaceExists(ctx, client.clientset, cassandraNamespace)
	if err != nil {
		return 0, err
	}

	if !exists {
		if verbose {
			fmt.Fprintf(header, "Cassandra was not found.\n")
		}

		return 0, nil
	}

	secret, err := client.clientset.CoreV1().Secrets(cassandraNamespace).Get(ctx, "k8ssandra-superuser", metav1.GetOptions{})
	if err != nil {
		return 0, err
	}

	username := secret.Data["username"]
	password := secret.Data["password"]
	output := &bytes.Buffer{}

	err = exec(
		client,
		cassandraNamespace,
		"k8ssandra-dc1-default-sts-0",
		"cassandra",
		fmt.Sprintf("nodetool -u %s -pw %s --host ::FFFF:127.0.0.1 status | grep --extended-regexp '^[UD][NLJM]\\s+'", username, password),
		output,
	)
	if err != nil {
		return 0, err
	}

	readyNodes := strings.Count(output.String(), "UN")
	totalNodes := strings.Count(output.String(), "\n")

	fmt.Fprintf(header, "%d of %d cassandra nodes are ready.\n", readyNodes, totalNodes)

	if readyNodes != totalNodes {
		fmt.Fprint(details, output.String())
		return 46, nil
	}

	return 0, nil
}

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

type cassandraStatus struct {
	found          bool
	total          int
	healthyCount   int
	unhealthyCount int
	details        string
}

func NewCassandraStatus(ctx context.Context, client *KubernetesClient) (status, error) {
	status := &cassandraStatus{}

	exists, err := namespaceExists(ctx, client.clientset, cassandraNamespace)
	if err != nil {
		return status, err
	}

	status.found = exists

	if !status.found {
		return status, nil
	}

	readyNodes, totalNodes, details, err := getCassandraNodeStatus(ctx, client)

	status.total = totalNodes
	status.healthyCount = readyNodes
	status.unhealthyCount = totalNodes - readyNodes
	status.details = details

	return status, nil
}

func (s *cassandraStatus) Summary(w io.Writer, verbose bool) error {
	if !verbose && !s.found {
		return nil
	}

	_, err := fmt.Fprintf(w, "%d of %d cassandra nodes are ready.\n", s.healthyCount, s.total)
	return err
}

func (s *cassandraStatus) Details(w io.Writer, verbose, colored bool) error {
	if s.unhealthyCount == 0 {
		return nil
	}

	_, err := fmt.Fprint(w, s.details)
	return err
}

func (s *cassandraStatus) ExitCode() int {
	if s.unhealthyCount > 0 {
		return 46
	}

	return 0
}

func getCassandraNodeStatus(ctx context.Context, client *KubernetesClient) (int, int, string, error) {
	username, password, err := getCasssandraCredentials(ctx, client)
	if err != nil {
		return 0, 0, "", err
	}

	outputBytes := &bytes.Buffer{}
	command := fmt.Sprintf("nodetool -u %s -pw %s --host ::FFFF:127.0.0.1 status | grep --extended-regexp '^[UD][NLJM]\\s+'", username, password)

	err = exec(
		client,
		cassandraNamespace,
		"k8ssandra-dc1-default-sts-0",
		"cassandra",
		command,
		outputBytes,
	)
	if err != nil {
		return 0, 0, "", err
	}

	output := outputBytes.String()
	readyNodes := strings.Count(output, "UN")
	totalNodes := strings.Count(output, "\n")

	return readyNodes, totalNodes, output, nil
}

func getCasssandraCredentials(ctx context.Context, client *KubernetesClient) (string, string, error) {
	secret, err := client.clientset.CoreV1().Secrets(cassandraNamespace).Get(ctx, "k8ssandra-superuser", metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}

	username := secret.Data["username"]
	password := secret.Data["password"]

	return string(username), string(password), nil
}

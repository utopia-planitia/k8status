package k8status

import (
	"context"
	"fmt"
	"io"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type nodesStatus struct {
	total          int
	healthyCount   int
	nodes          []v1.Node
	unhealthyCount int
}

func NewNodeStatus(ctx context.Context, client *KubernetesClient) (status, error) {
	nodesList, err := client.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nodesStatus{}, err
	}

	nodes := nodesList.Items

	status := nodesStatus{
		total:          len(nodes),
		healthyCount:   0,
		nodes:          []v1.Node{},
		unhealthyCount: 0,
	}
	status.add(nodes)

	return status, nil
}

func (s nodesStatus) Summary(w io.Writer, verbose bool) error {
	_, err := fmt.Fprintf(w, "%d of %d Nodes are up and healthy.\n", s.healthyCount, s.total)
	return err
}

func (s nodesStatus) Details(w io.Writer, verbose, colored bool) error {
	return s.toTable().Fprint(w, colored)
}

func (s nodesStatus) ExitCode() int {
	if s.unhealthyCount > 0 {
		return 43
	}

	return 0
}

func (s nodesStatus) toTable() Table {
	header := []string{"Node", "Status", "Messages"}

	rows := [][]string{}
	for _, node := range s.nodes {
		isReady, cordoned, messages := getNodeConditions(node)
		row := []string{node.Name, formatStatus(isReady, cordoned), strings.Join(messages, "; ")}
		rows = append(rows, row)
	}

	return Table{
		Header: header,
		Rows:   rows,
	}
}
func (s nodesStatus) add(nodes []v1.Node) {
	s.total += len(nodes)

	for _, item := range nodes {
		isReady, cordoned, _ := getNodeConditions(item)

		if nodeIsHealthy(isReady, cordoned) {
			s.healthyCount++
			continue
		}

		s.nodes = append(s.nodes, item)

		s.unhealthyCount++
	}
}

func getNodeConditions(node v1.Node) (bool, bool, []string) {
	messages := make([]string, 0)
	ready := false
	cordoned := node.Spec.Unschedulable

	for _, condition := range node.Status.Conditions {
		if condition.Type != v1.NodeReady {
			if condition.Status == v1.ConditionTrue {
				messages = append(messages, condition.Message)
			}
		}

		if condition.Type == v1.NodeReady {
			ready = condition.Status == v1.ConditionTrue
		}
	}

	return ready, cordoned, messages
}

func formatStatus(isReady bool, cordoned bool) string {
	statuses := []string{}
	if isReady {
		statuses = append(statuses, "Ready")
	}
	if cordoned {
		statuses = append(statuses, "Cordoned")
	}
	return strings.Join(statuses, ", ")
}

func nodeIsHealthy(isReady bool, cordoned bool) bool {
	return isReady && !cordoned
}

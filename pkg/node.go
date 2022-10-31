package k8status

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrNodeListIsNil error = errors.New("ErrNodeListIsNil")

type nodeTableView struct {
	name     string
	status   string
	messages string
}

func (c nodeTableView) header() []string {
	return []string{"Node", "Status", "Messages"}
}

func (c nodeTableView) row() []string {
	return []string{c.name, c.status, c.messages}
}

func PrintNodeStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose, colored bool) (int, error) {
	nodelist, err := client.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	return printNodeStatus(header, details, nodelist, verbose, colored)
}

func printNodeStatus(header io.Writer, details io.Writer, nodelist *v1.NodeList, verbose, colored bool) (int, error) {
	if nodelist == nil {
		return 0, ErrNodeListIsNil
	}

	stats := gatherNodesStats(nodelist)

	err := createAndWriteNodesTableInfo(header, details, stats, verbose, colored)
	if err != nil {
		return 0, err
	}

	exitCode := evaluateNodesStatus(stats)

	return exitCode, nil
}

func evaluateNodesStatus(stats *nodeStats) (exitCode int) {
	exitCode = 0

	if stats.foundUnhealthyNode {
		return 45
	}

	return exitCode
}

func createAndWriteNodesTableInfo(header io.Writer, details io.Writer, stats *nodeStats, verbose, colored bool) error {

	table, err := CreateTable(details, nodeTableView{}.header(), colored)
	if err != nil {
		return err
	}

	fmt.Fprintf(header, "%d of %d Nodes are up and healthy.\n", stats.healthyNodes, stats.nodesTotal)

	if verbose {
		if len(stats.tableData) != 0 {
			RenderTable(table, stats.tableData)
		}
	}

	return nil
}

type nodeStats struct {
	nodesTotal         int
	healthyNodes       int
	tableData          [][]string
	foundUnhealthyNode bool
}

func gatherNodesStats(nodelist *v1.NodeList) *nodeStats {
	foundUnhealthyNode := false

	healthy := 0
	tableData := [][]string{}

	for _, item := range nodelist.Items {
		isReady, cordoned, messages := getNodeConditions(item)

		if nodeIsHealthy(isReady, cordoned) {
			healthy++
		} else {
			tv := nodeTableView{item.Name, formatStatus(isReady, cordoned), strings.Join(messages, "; ")}
			tableData = append(tableData, tv.row())

			if isCiOrLabNamespace(item.Namespace) {
				continue
			}
			foundUnhealthyNode = true
		}
	}

	stats := nodeStats{
		nodesTotal:         len(nodelist.Items),
		healthyNodes:       healthy,
		tableData:          tableData,
		foundUnhealthyNode: foundUnhealthyNode,
	}

	return &stats
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
	return strings.Join(statuses, ",")
}

func nodeIsHealthy(isReady bool, cordoned bool) bool {
	return isReady && !cordoned
}

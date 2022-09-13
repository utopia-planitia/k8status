package k8status

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func PrintNodeStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	nodelist, err := getNodeList(ctx, client.clientset)
	if err != nil {
		return 0, err
	}

	up := 0
	count := len(nodelist.Items)
	exitCode := 0
	table, err := CreateTable(details, []string{"Node", "Status", "Messages"}, tablewriter.FgYellowColor)
	if err != nil {
		return 0, err
	}
	data := [][]string{}

	for _, node := range nodelist.Items {
		isReady, cordoned, messages := getNodeConditions(node)
		if isReady && !cordoned {
			up += 1
		} else {
			_ = messages
			//data = append(data, []string{node.Name, fmt.Sprintf("%t", isReady), fmt.Sprintf("%t", cordoned)})
			data = append(data, []string{node.Name, formatStatus(isReady, cordoned), strings.Join(messages, "; ")})
		}
	}

	fmt.Fprintf(header, "%d of %d Node are up and healthy.\n", up, count)

	if verbose {
		if len(data) != 0 {
			RenderTable(table, data)
		}
	}

	for _, node := range nodelist.Items {
		ready, cordoned, conditions := getNodeConditions(node)

		if !ready {
			exitCode = 45
		}

		if cordoned {
			exitCode = 45
		}

		if len(conditions) != 0 {
			for _, msg := range conditions {
				fmt.Fprintln(details, msg)
			}
			exitCode = 45
		}
	}

	return exitCode, nil
}

func getNodeList(ctx context.Context, clientset *kubernetes.Clientset) (*v1.NodeList, error) {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return nodes, nil
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

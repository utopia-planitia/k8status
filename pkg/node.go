package k8status

import (
	"context"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func PrintNodeStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose bool) (int, error) {
	nodelist, err := getNodeList(ctx, client.clientset)
	if err != nil {
		return 0, err
	}

	up := 0
	count := len(nodelist.Items)
	exitCode := 0

	for _, node := range nodelist.Items {
		isReady, cordoned, _ := getNodeConditions(node)
		if isReady && !cordoned {
			up += 1
		}
	}

	fmt.Fprintf(header, "%d of %d Node are up and healthy.\n", up, count)

	for _, node := range nodelist.Items {
		ready, cordoned, conditions := getNodeConditions(node)

		if !ready {
			fmt.Fprintf(details, "%s Node %s is not ready.\n", red("x"), node.Name)

			exitCode = 45
		}

		if cordoned {
			fmt.Fprintf(details, "%s Node %s is cordoned.\n", red("x"), node.Name)

			exitCode = 45
		}

		if len(conditions) != 0 {
			fmt.Fprintf(details, "%s Node %s has conditions:\n", red("x"), node.Name)

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

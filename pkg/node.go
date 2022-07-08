package k8status

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func PrintNodeStatus(ctx context.Context, restconfig *rest.Config, clientset *kubernetes.Clientset, verbose bool) (int, error) {
	nodelist, err := getNodeList(ctx, clientset)
	if err != nil {
		return 0, err
	}

	ready := 0
	count := len(nodelist.Items)

	for _, node := range nodelist.Items {
		isReady, _ := getNodeConditions(node)
		if isReady {
			ready += 1
		}
	}

	fmt.Printf("%d of %d Node are ready.\n", ready, count)

	for _, node := range nodelist.Items {
		isReady, conditions := getNodeConditions(node)

		if !isReady {
			fmt.Printf("%s Node %s is not ready.\n", red("x"), node.Name)
		} else if verbose {
			fmt.Printf("%s Node %s is ready.\n", green("✓"), node.Name)
		}

		for _, msg := range conditions {
			fmt.Println(msg)
		}
	}

	if ready != count {
		return 41, nil
	}

	return 0, nil
}

func getNodeList(ctx context.Context, clientset *kubernetes.Clientset) (*v1.NodeList, error) {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

func getNodeConditions(node v1.Node) (bool, []string) {
	messages := make([]string, 0)
	ready := false

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

	return ready, messages
}

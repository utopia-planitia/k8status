package k8status

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func PrintNodeStatus(ctx context.Context, clientset *kubernetes.Clientset) error {

	nodelist, err := getNodeList(ctx, clientset)
	if err != nil {
		return err
	}

	for _, node := range nodelist.Items {
		isReady, conditions := getNodeConditions(node)
		if isReady {
			fmt.Printf("%s Node %s is ready\n", green("✓"), node.Name)
		} else {
			fmt.Printf("%s Node %s is not ready\n", red("x"), node.Name)
		}
		for _, msg := range conditions {
			fmt.Println(msg)
		}
	}

	return nil
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

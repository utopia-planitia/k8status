package k8status

import (
	"context"
	"fmt"
	"io"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintStatefulsetStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose bool) (int, error) {
	statefulsets, err := client.clientset.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	_ = statefulsets
	if err != nil {
		return 0, err
	}

	healthy := 0
	total := 0

	for _, item := range statefulsets.Items {
		total++

		if item.Status.Replicas == item.Status.ReadyReplicas &&
			item.Status.Replicas == item.Status.CurrentReplicas &&
			item.Status.Replicas == item.Status.UpdatedReplicas {
			healthy++
		} else {
			if verbose {
				_, err = details.Write([]byte(fmt.Sprintf("In namespace \"%s\", statefulset \"%s\" should have \"%d\""+
					" replicas but has \"%d\" ready, \"%d\" current and \"%d\" updated \n",
					item.Namespace, item.Name, item.Status.Replicas, item.Status.ReadyReplicas,
					item.Status.CurrentReplicas, item.Status.UpdatedReplicas)))
				if err != nil {
					return 0, err
				}
			}
		}

	}

	fmt.Fprintf(header, "%d of %d statefulsets are healthy.\n", healthy, total)

	for _, item := range statefulsets.Items {

		if strings.Contains(item.Namespace, "ci") || strings.Contains(item.Namespace, "lab") {
			continue
		}

		deploymentHealthy := item.Status.Replicas == item.Status.ReadyReplicas &&
			item.Status.Replicas == item.Status.CurrentReplicas &&
			item.Status.Replicas == item.Status.UpdatedReplicas

		if !deploymentHealthy {
			return 50, nil
		}

	}

	return 0, err

}

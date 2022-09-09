package k8status

import (
	"context"
	"fmt"
	"io"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintDeploymentStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose bool) (int, error) {
	deployments, err := client.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	_ = deployments
	if err != nil {
		return 0, err
	}

	healthy := 0
	total := 0

	for _, item := range deployments.Items {
		total++

		if item.Status.Replicas == item.Status.UpdatedReplicas &&
			item.Status.Replicas == item.Status.ReadyReplicas &&
			item.Status.Replicas == item.Status.AvailableReplicas {
			healthy++
		} else {
			if verbose {
				_, err = details.Write([]byte(fmt.Sprintf("In namespace \"%s\", deployment \"%s\" should have \"%d\""+
					" replicas but has \"%d\" available, \"%d\" updated and \"%d\" ready \n",
					item.Namespace, item.Name, item.Status.Replicas, item.Status.AvailableReplicas,
					item.Status.UpdatedReplicas, item.Status.ReadyReplicas)))
				if err != nil {
					return 0, err
				}
			}
		}

	}

	fmt.Fprintf(header, "%d of %d deployments are healthy.\n", healthy, total)

	for _, item := range deployments.Items {

		if strings.Contains(item.Namespace, "ci") || strings.Contains(item.Namespace, "lab") {
			continue
		}

		deploymentHealthy := item.Status.Replicas == item.Status.UpdatedReplicas &&
			item.Status.Replicas == item.Status.ReadyReplicas &&
			item.Status.Replicas == item.Status.AvailableReplicas

		if !deploymentHealthy {
			return 48, nil
		}

	}

	return 0, err

}

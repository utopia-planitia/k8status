package k8status

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintDeploymentStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	deployments, err := client.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	_ = deployments
	if err != nil {
		return 0, err
	}

	healthy := 0
	total := 0
	table, err := CreateTable(details, []string{"Deployment", "Namespace", "Replicas", "Available", "Up-to-date", "Ready"}, tablewriter.FgWhiteColor)
	if err != nil {
		return 0, err
	}
	tableData := [][]string{}

	for _, item := range deployments.Items {
		total++

		if item.Status.Replicas == item.Status.UpdatedReplicas &&
			item.Status.Replicas == item.Status.ReadyReplicas &&
			item.Status.Replicas == item.Status.AvailableReplicas {
			healthy++
		} else {
			tableData = append(tableData, []string{item.Name, item.Namespace, fmt.Sprintf("%d", item.Status.Replicas),
				fmt.Sprintf("%d", item.Status.AvailableReplicas), fmt.Sprintf("%d", item.Status.UpdatedReplicas),
				fmt.Sprintf("%d", item.Status.ReadyReplicas)})
		}

	}

	fmt.Fprintf(header, "%d of %d deployments are healthy.\n", healthy, total)

	if verbose {
		if len(tableData) != 0 {
			RenderTable(table, tableData)
		}
	}

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

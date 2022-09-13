package k8status

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintStatefulsetStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	statefulsets, err := client.clientset.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	_ = statefulsets
	if err != nil {
		return 0, err
	}

	healthy := 0
	total := 0
	table, err := CreateTable(details, []string{"Statefulset", "Namespace", "Replicas", "Ready", "Current", "Updated"}, tablewriter.FgCyanColor)
	if err != nil {
		return 0, err
	}
	tableData := [][]string{}

	for _, item := range statefulsets.Items {
		total++

		if item.Status.Replicas == item.Status.ReadyReplicas &&
			item.Status.Replicas == item.Status.CurrentReplicas &&
			item.Status.Replicas == item.Status.UpdatedReplicas {
			healthy++
		} else {
			tableData = append(tableData, []string{item.Name, item.Namespace, fmt.Sprintf("%d", item.Status.Replicas),
				fmt.Sprintf("%d", item.Status.ReadyReplicas), fmt.Sprintf("%d", item.Status.CurrentReplicas),
				fmt.Sprintf("%d", item.Status.UpdatedReplicas)})
		}
	}

	fmt.Fprintf(header, "%d of %d statefulsets are healthy.\n", healthy, total)

	if verbose {
		if len(tableData) != 0 {
			RenderTable(table, tableData)
		}
	}

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

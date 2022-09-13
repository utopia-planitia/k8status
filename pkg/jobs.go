package k8status

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintJobStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	jobs, err := client.clientset.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	healthy := 0
	table, err := CreateTable(details, []string{"Job", "Namespace", "Active", "Completions", "Succeeded", "Failed"}, tablewriter.FgBlueColor)
	if err != nil {
		return 0, err
	}
	tableData := [][]string{}

	for _, item := range jobs.Items {
		if !isHealthy(item) {
			tableData = append(tableData, []string{item.Name, item.Namespace,
				fmt.Sprintf("%d", item.Status.Active), fmt.Sprintf("%d", *item.Spec.Completions),
				fmt.Sprintf("%d", item.Status.Succeeded), fmt.Sprintf("%d", item.Status.Failed)})
			continue
		}

		healthy++
	}

	fmt.Fprintf(header, "%d of %d jobs are completed.\n", healthy, len(jobs.Items))

	if verbose {
		if len(tableData) != 0 {
			RenderTable(table, tableData)
		}
	}

	for _, item := range jobs.Items {
		if strings.Contains(item.ObjectMeta.Namespace, "ci") || strings.Contains(item.ObjectMeta.Namespace, "lab") {
			continue
		}

		if isHealthy(item) {
			continue
		}

		return 49, nil
	}

	return 0, nil
}

func isHealthy(item v1.Job) bool {
	if item.Status.Active > 0 {
		return true
	}

	if *item.Spec.Completions == item.Status.Succeeded {
		return true
	}

	return false
}

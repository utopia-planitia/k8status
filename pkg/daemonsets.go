package k8status

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type daemonsetTableView struct {
	name      string
	namespace string
	scheduled string
	current   string
	ready     string
	updated   string
	available string
}

func (c daemonsetTableView) header() []string {
	return []string{"Daemonset", "Namespace", "Scheduled", "Current", "Ready", "Up-to-date", "Available"}
}

func (c daemonsetTableView) row() []string {
	return []string{c.name, c.namespace, c.scheduled, c.current, c.ready, c.updated, c.available}
}

func PrintDaemonsetStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	daemonsets, err := client.clientset.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	_ = daemonsets
	if err != nil {
		return 0, err
	}

	healthy := 0
	total := 0

	table, err := CreateTable(details, tableHeader(daemonsetTableView{}), tablewriter.FgYellowColor)
	if err != nil {
		return 0, err
	}
	tableData := [][]string{}

	for _, item := range daemonsets.Items {
		total++

		if item.Status.DesiredNumberScheduled == item.Status.CurrentNumberScheduled &&
			item.Status.DesiredNumberScheduled == item.Status.NumberReady &&
			item.Status.DesiredNumberScheduled == item.Status.UpdatedNumberScheduled &&
			item.Status.DesiredNumberScheduled == item.Status.NumberAvailable {
			healthy++
		} else {
			tableData = append(tableData, tableRow(daemonsetTableView{item.Name, item.Namespace, fmt.Sprintf("%d", item.Status.DesiredNumberScheduled),
				fmt.Sprintf("%d", item.Status.CurrentNumberScheduled), fmt.Sprintf("%d", item.Status.NumberReady),
				fmt.Sprintf("%d", item.Status.UpdatedNumberScheduled), fmt.Sprintf("%d", item.Status.NumberAvailable)}))

		}

	}

	fmt.Fprintf(header, "%d of %d daemonsets are healthy.\n", healthy, total)

	if verbose {
		if len(tableData) != 0 {
			RenderTable(table, tableData)
		}
	}

	for _, item := range daemonsets.Items {

		if strings.Contains(item.Namespace, "ci") || strings.Contains(item.Namespace, "lab") {
			continue
		}

		deploymentHealthy := item.Status.DesiredNumberScheduled == item.Status.CurrentNumberScheduled &&
			item.Status.DesiredNumberScheduled == item.Status.NumberReady &&
			item.Status.DesiredNumberScheduled == item.Status.UpdatedNumberScheduled &&
			item.Status.DesiredNumberScheduled == item.Status.NumberAvailable

		if !deploymentHealthy {
			return 51, nil
		}

	}

	return 0, err

}

package k8status

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrDaemonsetListIsNil error = errors.New("ErrDaemonsetListIsNil")

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

func printDaemonsetStatus(_ context.Context, header io.Writer, details colorWriter, daemonsets *appsv1.DaemonSetList, verbose bool) (int, error) {
	if daemonsets == nil {
		return 0, ErrDaemonsetListIsNil
	}

	stats := gatherDaemonsetsStats(daemonsets)

	err := createAndWriteTableInfo(header, details, stats, verbose)
	if err != nil {
		return 0, err
	}

	exitCode := evaluateDaemonsetsStatus(stats)

	return exitCode, nil
}

func evaluateDaemonsetsStatus(stats *daemonsetStats) (exitCode int) {
	exitCode = 0

	if stats.foundDaemonset {
		return 51
	}

	return exitCode
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

	// for _, item := range daemonsets.Items {

	// 	if strings.Contains(item.Namespace, "ci") || strings.Contains(item.Namespace, "lab") {
	// 		continue
	// 	}

	// 	deploymentHealthy := item.Status.DesiredNumberScheduled == item.Status.CurrentNumberScheduled &&
	// 		item.Status.DesiredNumberScheduled == item.Status.NumberReady &&
	// 		item.Status.DesiredNumberScheduled == item.Status.UpdatedNumberScheduled &&
	// 		item.Status.DesiredNumberScheduled == item.Status.NumberAvailable

	// 	if !deploymentHealthy {
	// 		return 51, nil
	// 	}

	// }
	exitCode := evaluateDaemonsetsStatus(stats)

	return exitCode, err

}

func gatherDaemonsetsStats(daemonsets *appsv1.DaemonSetList) *daemonsetStats {

	healthy := 0
	tableData := [][]string{}

	for _, item := range daemonsets.Items {

		if strings.Contains(item.Namespace, "ci") || strings.Contains(item.Namespace, "lab") {
			continue
		}

		daemonsetHealthy := item.Status.DesiredNumberScheduled == item.Status.CurrentNumberScheduled &&
			item.Status.DesiredNumberScheduled == item.Status.NumberReady &&
			item.Status.DesiredNumberScheduled == item.Status.UpdatedNumberScheduled &&
			item.Status.DesiredNumberScheduled == item.Status.NumberAvailable

		if daemonsetHealthy {
			healthy++
		}
	}

	stats := daemonsetStats{
		total:       len(daemonsets.Items),
		healthySets: healthy,
		tableData:   tableData,
	}

	return &stats
}

type daemonsetStats struct {
	total       int
	healthySets int
	tableData   [][]string
}

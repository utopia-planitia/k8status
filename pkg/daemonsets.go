package k8status

import (
	"context"
	"errors"
	"fmt"
	"io"

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

func PrintDaemonsetStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	daemonsets, err := client.clientset.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	_ = daemonsets
	if err != nil {
		return 0, err
	}

	return printDaemonsetStatus(ctx, header, details, daemonsets, verbose)
}

func printDaemonsetStatus(_ context.Context, header io.Writer, details colorWriter, daemonsets *appsv1.DaemonSetList, verbose bool) (int, error) {
	if daemonsets == nil {
		return 0, ErrDaemonsetListIsNil
	}

	stats := gatherDaemonsetsStats(daemonsets)

	err := createAndWriteDaemonsetsTableInfo(header, details, stats, verbose)
	if err != nil {
		return 0, err
	}

	exitCode := evaluateDaemonsetsStatus(stats)

	return exitCode, nil
}

func evaluateDaemonsetsStatus(stats *daemonsetStats) (exitCode int) {
	exitCode = 0

	if stats.foundUnhealthyDaemonset {
		return 51
	}

	return exitCode
}

func createAndWriteDaemonsetsTableInfo(header io.Writer, details colorWriter, stats *daemonsetStats, verbose bool) error {

	table, err := CreateTable(details, tableHeader(daemonsetTableView{}), tablewriter.FgYellowColor)
	if err != nil {
		return err
	}

	fmt.Fprintf(header, "%d of %d daemonsets are healthy.\n", stats.healthySets, stats.setsTotal)

	if verbose {
		if len(stats.tableData) != 0 {
			RenderTable(table, stats.tableData) //"renders" (not really) by writing into the details writer
		}
	}

	return nil
}

type daemonsetStats struct {
	setsTotal               int
	healthySets             int
	tableData               [][]string
	foundUnhealthyDaemonset bool
}

func gatherDaemonsetsStats(daemonsets *appsv1.DaemonSetList) *daemonsetStats {
	foundUnhealthyDaemonset := false

	healthy := 0
	tableData := [][]string{}

	for _, item := range daemonsets.Items {

		if daemonsetIsHealthy(item) {
			healthy++
		} else {
			tableData = append(tableData, tableRow(daemonsetTableView{item.Name, item.Namespace, fmt.Sprintf("%d", item.Status.DesiredNumberScheduled),
				fmt.Sprintf("%d", item.Status.CurrentNumberScheduled), fmt.Sprintf("%d", item.Status.NumberReady),
				fmt.Sprintf("%d", item.Status.UpdatedNumberScheduled), fmt.Sprintf("%d", item.Status.NumberAvailable)}))

			if isCiOrLabNamespace(item.Namespace) {
				continue
			}
			foundUnhealthyDaemonset = true
		}
	}

	stats := daemonsetStats{
		setsTotal:               len(daemonsets.Items),
		healthySets:             healthy,
		tableData:               tableData,
		foundUnhealthyDaemonset: foundUnhealthyDaemonset,
	}

	return &stats
}

func daemonsetIsHealthy(item appsv1.DaemonSet) bool {
	return item.Status.DesiredNumberScheduled == item.Status.CurrentNumberScheduled &&
		item.Status.DesiredNumberScheduled == item.Status.NumberReady &&
		item.Status.DesiredNumberScheduled == item.Status.UpdatedNumberScheduled &&
		item.Status.DesiredNumberScheduled == item.Status.NumberAvailable
}

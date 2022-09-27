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

var ErrStatefulsetListIsNil error = errors.New("ErrStatefulsetListIsNil")

type statefulsetTableView struct {
	name      string
	namespace string
	replicas  string
	ready     string
	current   string
	updated   string
}

func (c statefulsetTableView) header() []string {
	return []string{"Statefulset", "Namespace", "Replicas", "Ready", "Current", "Updated"}
}

func (c statefulsetTableView) row() []string {
	return []string{c.name, c.namespace, c.replicas, c.ready, c.current, c.updated}
}

func PrintStatefulsetStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	statefulsets, err := client.clientset.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	_ = statefulsets
	if err != nil {
		return 0, err
	}

	return printStatefulsetStatus(ctx, header, details, statefulsets, verbose)

	// healthy := 0
	// total := 0
	// table, err := CreateTable(details, []string{"Statefulset", "Namespace", "Replicas", "Ready", "Current", "Updated"}, tablewriter.FgCyanColor)
	// if err != nil {
	// 	return 0, err
	// }
	// tableData := [][]string{}

	// for _, item := range statefulsets.Items {
	// 	total++

	// 	if item.Status.Replicas == item.Status.ReadyReplicas &&
	// 		item.Status.Replicas == item.Status.CurrentReplicas &&
	// 		item.Status.Replicas == item.Status.UpdatedReplicas {
	// 		healthy++
	// 	} else {
	// 		tableData = append(tableData, []string{item.Name, item.Namespace, fmt.Sprintf("%d", item.Status.Replicas),
	// 			fmt.Sprintf("%d", item.Status.ReadyReplicas), fmt.Sprintf("%d", item.Status.CurrentReplicas),
	// 			fmt.Sprintf("%d", item.Status.UpdatedReplicas)})
	// 	}
	// }

	// fmt.Fprintf(header, "%d of %d statefulsets are healthy.\n", healthy, total)

	// if verbose {
	// 	if len(tableData) != 0 {
	// 		RenderTable(table, tableData)
	// 	}
	// }

	// for _, item := range statefulsets.Items {

	// 	if strings.Contains(item.Namespace, "ci") || strings.Contains(item.Namespace, "lab") {
	// 		continue
	// 	}

	// 	deploymentHealthy := item.Status.Replicas == item.Status.ReadyReplicas &&
	// 		item.Status.Replicas == item.Status.CurrentReplicas &&
	// 		item.Status.Replicas == item.Status.UpdatedReplicas

	// 	if !deploymentHealthy {
	// 		return 50, nil
	// 	}

	// }

	// return 0, err

}

func printStatefulsetStatus(_ context.Context, header io.Writer, details colorWriter, statefulsets *appsv1.StatefulSetList, verbose bool) (int, error) {
	if statefulsets == nil {
		return 0, ErrStatefulsetListIsNil
	}

	stats := gatherStatefulsetsStats(statefulsets)

	err := createAndWriteStatefulsetsTableInfo(header, details, stats, verbose)
	if err != nil {
		return 0, err
	}

	exitCode := evaluateStatefulsetsStatus(stats)

	return exitCode, nil
}

func evaluateStatefulsetsStatus(stats *statefulsetStats) (exitCode int) {
	exitCode = 0

	if stats.foundUnhealthyStatefulset {
		return 50
	}

	return exitCode
}

func createAndWriteStatefulsetsTableInfo(header io.Writer, details colorWriter, stats *statefulsetStats, verbose bool) error {

	table, err := CreateTable(details, tableHeader(statefulsetTableView{}), tablewriter.FgYellowColor)
	if err != nil {
		return err
	}

	fmt.Fprintf(header, "%d of %d statefulsets are healthy.\n", stats.healthySets, stats.setsTotal)

	if verbose {
		if len(stats.tableData) != 0 {
			RenderTable(table, stats.tableData) //"renders" (not really) by writing into the details writer
		}
	}

	return nil
}

type statefulsetStats struct {
	setsTotal                 int
	healthySets               int
	tableData                 [][]string
	foundUnhealthyStatefulset bool
}

func gatherStatefulsetsStats(statefulsets *appsv1.StatefulSetList) *statefulsetStats {
	foundUnhealthyStatefulset := false

	healthy := 0
	tableData := [][]string{}

	for _, item := range statefulsets.Items {

		if statefulsetIsHealthy(item) {
			healthy++
		} else {
			tableData = append(tableData, tableRow(statefulsetTableView{item.Name, item.Namespace, fmt.Sprintf("%d", item.Status.Replicas),
				fmt.Sprintf("%d", item.Status.ReadyReplicas), fmt.Sprintf("%d", item.Status.CurrentReplicas),
				fmt.Sprintf("%d", item.Status.UpdatedReplicas)}))

			if strings.Contains(item.Namespace, "ci") || strings.Contains(item.Namespace, "lab") {
				continue
			}
			foundUnhealthyStatefulset = true
		}
	}

	stats := statefulsetStats{
		setsTotal:                 len(statefulsets.Items),
		healthySets:               healthy,
		tableData:                 tableData,
		foundUnhealthyStatefulset: foundUnhealthyStatefulset,
	}

	return &stats
}

func statefulsetIsHealthy(item appsv1.StatefulSet) bool {
	return item.Status.Replicas == item.Status.ReadyReplicas &&
		item.Status.Replicas == item.Status.CurrentReplicas &&
		item.Status.Replicas == item.Status.UpdatedReplicas
}

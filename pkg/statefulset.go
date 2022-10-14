package k8status

import (
	"context"
	"errors"
	"fmt"
	"io"

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

func PrintStatefulsetStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose, colored bool) (int, error) {
	statefulsets, err := client.clientset.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	_ = statefulsets
	if err != nil {
		return 0, err
	}

	return printStatefulsetStatus(header, details, statefulsets, verbose, colored)
}

func printStatefulsetStatus(header io.Writer, details io.Writer, statefulsets *appsv1.StatefulSetList, verbose, colored bool) (int, error) {
	if statefulsets == nil {
		return 0, ErrStatefulsetListIsNil
	}

	stats := gatherStatefulsetsStats(statefulsets)

	err := createAndWriteStatefulsetsTableInfo(header, details, stats, verbose, colored)
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

func createAndWriteStatefulsetsTableInfo(header io.Writer, details io.Writer, stats *statefulsetStats, verbose, colored bool) error {
	table, err := CreateTable(details, tableHeader(statefulsetTableView{}), colored)
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

			if isCiOrLabNamespace(item.Namespace) {
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

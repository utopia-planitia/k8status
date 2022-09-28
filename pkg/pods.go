package k8status

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrPodListIsNil error = errors.New("ErrPodListIsNil")

type podTableView struct {
	name      string
	namespace string
	phase     string
	ready     string
	expected  string
}

func (c podTableView) header() []string {
	return []string{"Pod", "Namespace", "Phase", "Containers Ready", "Containers Expected"}
}

func (c podTableView) row() []string {
	return []string{c.name, c.namespace, c.phase, c.ready, c.expected}
}

func PrintPodStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	pods, err := client.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	return printPodStatus(ctx, header, details, pods, verbose)

	// healthy := 0
	// total := 0
	// table, err := CreateTable(details, []string{"Pod", "Namespace", "Phase", "Containers Ready", "Containers Expected"}, tablewriter.FgRedColor)
	// if err != nil {
	// 	return 0, err
	// }
	// tableData := [][]string{}

	// for _, item := range pods.Items {
	// 	if item.Status.Phase == v1.PodSucceeded || item.Status.Phase == v1.PodFailed {
	// 		continue
	// 	}

	// 	total++

	// 	containerReady := 0
	// 	for _, containerStatus := range item.Status.ContainerStatuses {
	// 		if containerStatus.Ready {
	// 			containerReady++
	// 		}
	// 	}

	// 	if len(item.Spec.Containers) == containerReady {
	// 		healthy++
	// 	} else {
	// 		tableData = append(tableData, []string{item.Name, item.Namespace, string(item.Status.Phase),
	// 			fmt.Sprintf("%d", containerReady), fmt.Sprintf("%d", len(item.Spec.Containers))})
	// 	}
	// }

	// fmt.Fprintf(header, "%d of %d pods are running.\n", healthy, total)

	// if verbose {
	// 	if len(tableData) != 0 {
	// 		RenderTable(table, tableData)
	// 	}
	// }

	// for _, item := range pods.Items {
	// 	if item.Status.Phase == v1.PodSucceeded {
	// 		continue
	// 	}

	// 	if strings.Contains(item.ObjectMeta.Namespace, "ci") || strings.Contains(item.ObjectMeta.Namespace, "lab") {
	// 		continue
	// 	}

	// 	containerReady := 0
	// 	for _, containerStatus := range item.Status.ContainerStatuses {
	// 		if containerStatus.Ready {
	// 			containerReady++
	// 		}
	// 	}

	// 	if len(item.Spec.Containers) != containerReady {
	// 		return 45, nil
	// 	}
	// }

	// return 0, err
}

func printPodStatus(_ context.Context, header io.Writer, details colorWriter, pods *v1.PodList, verbose bool) (int, error) {
	if pods == nil {
		return 0, ErrPodListIsNil
	}

	stats := gatherPodsStats(pods)

	err := createAndWritePodsTableInfo(header, details, stats, verbose)
	if err != nil {
		return 0, err
	}

	exitCode := evaluatePodsStatus(stats)

	return exitCode, nil
}

func evaluatePodsStatus(stats *podsStats) (exitCode int) {
	exitCode = 0

	if stats.foundUnhealthyPod {
		return 45
	}

	return exitCode
}

func createAndWritePodsTableInfo(header io.Writer, details colorWriter, stats *podsStats, verbose bool) error {

	table, err := CreateTable(details, tableHeader(podTableView{}), tablewriter.FgYellowColor)
	if err != nil {
		return err
	}

	fmt.Fprintf(header, "%d of %d pods are healthy.\n", stats.healthyPods, stats.podsTotal)

	if verbose {
		if len(stats.tableData) != 0 {
			RenderTable(table, stats.tableData) //"renders" (not really) by writing into the details writer
		}
	}

	return nil
}

type podsStats struct {
	podsTotal         int
	healthyPods       int
	tableData         [][]string
	foundUnhealthyPod bool
}

func gatherPodsStats(pods *v1.PodList) *podsStats {
	foundUnhealthyPod := false

	healthy := 0
	total := 0
	tableData := [][]string{}

	for _, item := range pods.Items {

		// TODO: Two lines commented below:
		// The first one was commented because there was a discrepancy between the result of k8status and status.sh
		// The secont one was also added because there was still a discrepancy between the result of k8status and status.sh
		//if item.Status.Phase == v1.PodSucceeded || item.Status.Phase == v1.PodFailed {
		if item.Status.Phase == v1.PodSucceeded {
			healthy++ // was not here
			continue
		}
		total++

		if podIsHealthy(item) {
			healthy++
		} else {
			tableData = append(tableData, tableRow(podTableView{item.Name, item.Namespace, string(item.Status.Phase),
				fmt.Sprintf("%d", getReadyContainers(item)), fmt.Sprintf("%d", len(item.Spec.Containers))}))

			if isCiOrLabNamespace(item.Namespace) {
				continue
			}
			foundUnhealthyPod = true
		}
	}

	stats := podsStats{
		podsTotal:         len(pods.Items),
		healthyPods:       healthy,
		tableData:         tableData,
		foundUnhealthyPod: foundUnhealthyPod,
	}

	return &stats
}

func podIsHealthy(item v1.Pod) bool {
	return len(item.Spec.Containers) == getReadyContainers(item)
}

func getReadyContainers(item v1.Pod) int {
	containerReady := 0
	for _, containerStatus := range item.Status.ContainerStatuses {
		if containerStatus.Ready {
			containerReady++
		}
	}
	return containerReady
}

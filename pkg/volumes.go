package k8status

import (
	"context"
	"errors"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrVolumeListIsNil error = errors.New("ErrVolumeListIsNil")

type volumeTableView struct {
	name      string
	namespace string
	phase     string
}

func (c volumeTableView) header() []string {
	return []string{"Volume", "Namespace", "Phase"}
}

func (c volumeTableView) row() []string {
	return []string{c.name, c.namespace, c.phase}
}

func PrintVolumeStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose, colored bool) (int, error) {
	pvs, err := client.clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	return printVolumeStatus(header, details, pvs, verbose, colored)
}

func printVolumeStatus(header io.Writer, details io.Writer, pvs *v1.PersistentVolumeList, verbose, colored bool) (int, error) {
	if pvs == nil {
		return 0, ErrVolumeListIsNil
	}

	stats := gatherVolumesStats(pvs)

	err := createAndWriteVolumesTableInfo(header, details, stats, verbose, colored)
	if err != nil {
		return 0, err
	}

	exitCode := evaluateVolumesStatus(stats)

	return exitCode, nil
}

func evaluateVolumesStatus(stats *volumeStats) (exitCode int) {
	exitCode = 0

	if stats.foundUnhealthyVolume {
		return 42
	}

	return exitCode
}

func createAndWriteVolumesTableInfo(header io.Writer, details io.Writer, stats *volumeStats, verbose, colored bool) error {
	table, err := CreateTable(details, volumeTableView{}.header(), colored)
	if err != nil {
		return err
	}

	fmt.Fprintf(header, "%d of %d volumes are bound or available.\n", stats.healthyVolumes, stats.volumesTotal)

	if verbose {
		if len(stats.tableData) != 0 {
			RenderTable(table, stats.tableData) //"renders" (not really) by writing into the details writer
		}
	}

	return nil
}

type volumeStats struct {
	volumesTotal         int
	healthyVolumes       int
	tableData            [][]string
	foundUnhealthyVolume bool
}

func gatherVolumesStats(pvs *v1.PersistentVolumeList) *volumeStats {
	foundUnhealthyVolume := false

	healthy := 0
	tableData := [][]string{}

	for _, item := range pvs.Items {

		if volumeIsHealthy(item) {
			healthy++
		} else {
			tv := volumeTableView{item.Name, item.Namespace, string(item.Status.Phase)}
			tableData = append(tableData, tv.row())

			if isCiOrLabNamespace(item.Namespace) {
				continue
			}
			foundUnhealthyVolume = true
		}
	}

	stats := volumeStats{
		volumesTotal:         len(pvs.Items),
		healthyVolumes:       healthy,
		tableData:            tableData,
		foundUnhealthyVolume: foundUnhealthyVolume,
	}

	return &stats
}

func volumeIsHealthy(item v1.PersistentVolume) bool {
	return item.Status.Phase == v1.VolumeBound || item.Status.Phase == v1.VolumeAvailable
}

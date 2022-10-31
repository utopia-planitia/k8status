package k8status

import (
	"context"
	"errors"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrVolumeClaimsListIsNil error = errors.New("ErrVolumeClaimsListIsNil")

type volumeClaimsTableView struct {
	name      string
	namespace string
	phase     string
}

func (c volumeClaimsTableView) header() []string {
	return []string{"Volume Claim", "Namespace", "Phase"}
}

func (c volumeClaimsTableView) row() []string {
	return []string{c.name, c.namespace, c.phase}
}

func PrintVolumeClaimStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose, colored bool) (int, error) {
	pvcs, err := client.clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	return printVolumeClaimStatus(header, details, pvcs, verbose, colored)
}

func printVolumeClaimStatus(header io.Writer, details io.Writer, pvcs *v1.PersistentVolumeClaimList, verbose, colored bool) (int, error) {
	if pvcs == nil {
		return 0, ErrVolumeClaimsListIsNil
	}

	stats := gatherVolumeClaimsStats(pvcs)

	err := createAndWriteVolumeClaimsTableInfo(header, details, stats, verbose, colored)
	if err != nil {
		return 0, err
	}

	exitCode := evaluateVolumeClaimsStatus(stats)

	return exitCode, nil
}

func evaluateVolumeClaimsStatus(stats *volumeClaimsStats) (exitCode int) {
	exitCode = 0

	if stats.foundUnhealthyVolumeClaim {
		return 43
	}

	return exitCode
}

func createAndWriteVolumeClaimsTableInfo(header io.Writer, details io.Writer, stats *volumeClaimsStats, verbose, colored bool) error {

	table, err := CreateTable(details, volumeClaimsTableView{}.header(), colored)
	if err != nil {
		return err
	}

	fmt.Fprintf(header, "%d of %d volume claims are bound.\n", stats.healthyVolumeClaims, stats.volumeClaimsTotal)

	if verbose {
		if len(stats.tableData) != 0 {
			RenderTable(table, stats.tableData) // "renders" (not really) by writing into the details writer
		}
	}

	return nil
}

type volumeClaimsStats struct {
	volumeClaimsTotal         int
	healthyVolumeClaims       int
	tableData                 [][]string
	foundUnhealthyVolumeClaim bool
}

func gatherVolumeClaimsStats(pvcs *v1.PersistentVolumeClaimList) *volumeClaimsStats {
	foundUnhealthyVolumeClaim := false

	healthy := 0
	tableData := [][]string{}

	for _, item := range pvcs.Items {
		if volumeClaimIsHealthy(item) {
			healthy++
			continue
		}
		tv := volumeClaimsTableView{item.Name, item.Namespace, string(item.Status.Phase)}
		tableData = append(tableData, tv.row())

		if !isCiOrLabNamespace(item.Namespace) {
			foundUnhealthyVolumeClaim = true
		}
	}

	stats := volumeClaimsStats{
		volumeClaimsTotal:         len(pvcs.Items),
		healthyVolumeClaims:       healthy,
		tableData:                 tableData,
		foundUnhealthyVolumeClaim: foundUnhealthyVolumeClaim,
	}

	return &stats
}

func volumeClaimIsHealthy(item v1.PersistentVolumeClaim) bool {
	return item.Status.Phase == v1.ClaimBound
}

package k8status

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
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

func PrintVolumeClaimStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	pvcs, err := client.clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	return printVolumeClaimStatus(ctx, header, details, pvcs, verbose)

	// healthy := 0
	// for _, item := range pvcs.Items {
	// 	if item.Status.Phase == v1.ClaimBound {
	// 		healthy++
	// 	}
	// }

	// fmt.Fprintf(header, "%d of %d volume claims are bound.\n", healthy, len(pvcs.Items))

	// if len(pvcs.Items) != healthy {
	// 	for _, item := range pvcs.Items {
	// 		if item.Status.Phase != v1.ClaimBound {
	// 			fmt.Fprintf(details, "Persistent Volume Claim %s in namespace %s is in phase %s\n", item.Name, item.Namespace, item.Status.Phase)
	// 		}
	// 	}

	// 	return 43, nil
	// }

	// return 0, nil
}

func printVolumeClaimStatus(_ context.Context, header io.Writer, details colorWriter, pvcs *v1.PersistentVolumeClaimList, verbose bool) (int, error) {
	if pvcs == nil {
		return 0, ErrVolumeClaimsListIsNil
	}

	stats := gatherVolumeClaimsStats(pvcs)

	err := createAndWriteVolumeClaimsTableInfo(header, details, stats, verbose)
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

func createAndWriteVolumeClaimsTableInfo(header io.Writer, details colorWriter, stats *volumeClaimsStats, verbose bool) error {

	table, err := CreateTable(details, tableHeader(volumeClaimsTableView{}), tablewriter.FgYellowColor)
	if err != nil {
		return err
	}

	fmt.Fprintf(header, "%d of %d volume claims are bound.\n", stats.healthyVolumeClaims, stats.volumeClaimsTotal)

	if verbose {
		if len(stats.tableData) != 0 {
			RenderTable(table, stats.tableData) //"renders" (not really) by writing into the details writer
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
		} else {
			tableData = append(tableData, tableRow(volumeClaimsTableView{item.Name, item.Namespace, string(item.Status.Phase)}))

			if strings.Contains(item.Namespace, "ci") || strings.Contains(item.Namespace, "lab") {
				continue
			}
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

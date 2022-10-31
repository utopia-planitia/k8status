package k8status

import (
	"context"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type volumeClaimsStats struct {
	total          int
	healthyCount   int
	claims         []v1.PersistentVolumeClaim
	unhealthyCount int
}

func (s volumeClaimsStats) addClaim(c v1.PersistentVolumeClaim) {
	s.claims = append(s.claims, c)
}

func (s volumeClaimsStats) summary(w io.Writer) {
	fmt.Fprintf(w, "%d of %d volume claims are bound.\n", s.healthyCount, s.total)
}

func (s volumeClaimsStats) toTable() Table {
	header := []string{"Volume Claim", "Namespace", "Phase"}

	rows := [][]string{}
	for _, c := range s.claims {
		row := []string{c.Name, c.Namespace, string(c.Status.Phase)}
		rows = append(rows, row)
	}

	return Table{
		Header: header,
		Rows:   rows,
	}
}

func PrintVolumeClaimStatus(
	ctx context.Context,
	header io.Writer,
	details io.Writer,
	client *KubernetesClient,
	verbose,
	colored bool,
) (int, error) {
	pvcs, err := client.clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	stats := statsFromList(pvcs)

	err = printStats(stats, header, details, verbose, colored)
	if err != nil {
		return 0, err
	}

	exitCode := stats.ExitCode()

	return exitCode, nil
}

func (s volumeClaimsStats) ExitCode() int {
	if s.unhealthyCount > 0 {
		return 43
	}

	return 0
}

func statsFromList(pvcs *v1.PersistentVolumeClaimList) volumeClaimsStats {
	stats := volumeClaimsStats{
		total:          len(pvcs.Items),
		healthyCount:   0,
		claims:         []v1.PersistentVolumeClaim{},
		unhealthyCount: 0,
	}

	for _, item := range pvcs.Items {
		if volumeClaimIsHealthy(item) {
			stats.healthyCount++
			continue
		}

		stats.addClaim(item)

		if !isCiOrLabNamespace(item.Namespace) {
			stats.unhealthyCount++
		}
	}

	return stats
}

func volumeClaimIsHealthy(item v1.PersistentVolumeClaim) bool {
	return item.Status.Phase == v1.ClaimBound
}

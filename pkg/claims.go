package k8status

import (
	"context"
	"fmt"
	"io"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type volumeClaimsStatus struct {
	total     int
	ignored   int
	healthy   int
	claims    []v1.PersistentVolumeClaim
	unhealthy int
}

func NewVolumeClaimsStatus(ctx context.Context, client *KubernetesClient) (status, error) {
	pvcsList, err := client.clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pvcs := pvcsList.Items

	status := &volumeClaimsStatus{
		claims: []v1.PersistentVolumeClaim{},
	}
	status.add(pvcs)

	return status, nil
}

func (s *volumeClaimsStatus) Summary(w io.Writer) error {
	return printSummaryWithIgnored(w, "%d of %d volume claims are bound.\n", s.ignored, s.healthy, s.total)
}

func printSummaryWithIgnored(w io.Writer, phrase string, ignored int, args ...any) error {
	if ignored != 0 {
		phrase = strings.ReplaceAll(phrase, ".\n", " (%d ignored).\n")
		args = append(args, ignored)
	}

	_, err := fmt.Fprintf(w, phrase, args...)
	return err
}

func (s *volumeClaimsStatus) Details(w io.Writer, colored bool) error {
	return s.toTable().Fprint(w, colored)
}

func (s *volumeClaimsStatus) ExitCode() int {
	if s.unhealthy > s.ignored {
		return 43
	}

	return 0
}

func (s *volumeClaimsStatus) toTable() Table {
	header := []string{"Namespace", "Volume Claim", "Phase"}

	rows := [][]string{}
	for _, item := range s.claims {
		row := []string{item.Namespace, item.Name, string(item.Status.Phase)}
		rows = append(rows, row)
	}

	return Table{
		Header: header,
		Rows:   rows,
	}
}

func (s *volumeClaimsStatus) add(pvcs []v1.PersistentVolumeClaim) {
	s.total += len(pvcs)

	for _, item := range pvcs {
		if volumeClaimIsHealthy(item) {
			s.healthy++
			continue
		}

		if isCiOrLabNamespace(item.Namespace) {
			s.ignored++
		}

		s.claims = append(s.claims, item)
		s.unhealthy++
	}
}

func volumeClaimIsHealthy(item v1.PersistentVolumeClaim) bool {
	return item.Status.Phase == v1.ClaimBound
}

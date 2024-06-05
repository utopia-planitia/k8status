package k8status

import (
	"context"
	"fmt"
	"io"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type daemonsetsStatus struct {
	total      int
	ignored    int
	healthy    int
	daemonSets []appsv1.DaemonSet
	unhealthy  int
}

func NewDaemonsetsStatus(ctx context.Context, client *KubernetesClient) (status, error) {
	daemonsetsList, err := client.clientset.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	daemonsets := daemonsetsList.Items

	status := &daemonsetsStatus{
		daemonSets: []appsv1.DaemonSet{},
	}
	status.add(daemonsets)

	return status, nil
}

func (s *daemonsetsStatus) Summary(w io.Writer) error {
	return printSummaryWithIgnored(w, "%d of %d daemonsets are healthy.\n", s.ignored, s.healthy, s.total)
}

func (s *daemonsetsStatus) Details(w io.Writer, colored bool) error {
	return s.toTable().Fprint(w, colored)
}

func (s *daemonsetsStatus) ExitCode() int {
	if s.unhealthy > s.ignored {
		return 51
	}

	return 0
}

func (s *daemonsetsStatus) toTable() Table {
	header := []string{"Namespace", "Daemonset", "Scheduled", "Current", "Ready", "Up-to-date", "Available"}

	rows := [][]string{}
	for _, item := range s.daemonSets {
		row := []string{
			item.Namespace,
			item.Name,
			fmt.Sprintf("%d", item.Status.DesiredNumberScheduled),
			fmt.Sprintf("%d", item.Status.CurrentNumberScheduled),
			fmt.Sprintf("%d", item.Status.NumberReady),
			fmt.Sprintf("%d", item.Status.UpdatedNumberScheduled),
			fmt.Sprintf("%d", item.Status.NumberAvailable),
		}
		rows = append(rows, row)
	}

	return Table{
		Header: header,
		Rows:   rows,
	}
}

func (s *daemonsetsStatus) add(pvcs []appsv1.DaemonSet) {
	s.total += len(pvcs)

	for _, item := range pvcs {
		if daemonsetIsHealthy(item) {
			s.healthy++
			continue
		}

		if isCiOrLabNamespace(item.Namespace) {
			s.ignored++
		}

		s.daemonSets = append(s.daemonSets, item)
		s.unhealthy++
	}
}

func daemonsetIsHealthy(item appsv1.DaemonSet) bool {
	return item.Status.DesiredNumberScheduled == item.Status.CurrentNumberScheduled &&
		item.Status.DesiredNumberScheduled == item.Status.NumberReady &&
		item.Status.DesiredNumberScheduled == item.Status.UpdatedNumberScheduled &&
		item.Status.DesiredNumberScheduled == item.Status.NumberAvailable
}

package k8status

import (
	"context"
	"fmt"
	"io"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type statefulsetsStatus struct {
	total        int
	ignored      int
	healthy      int
	statefulsets []appsv1.StatefulSet
	unhealthy    int
}

func NewStatefulsetsStatus(ctx context.Context, client *KubernetesClient) (status, error) {
	statefulsetsList, err := client.clientset.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	statefulsets := statefulsetsList.Items

	status := &statefulsetsStatus{
		statefulsets: []appsv1.StatefulSet{},
	}
	status.add(statefulsets)

	return status, nil
}

func (s *statefulsetsStatus) Summary(w io.Writer) error {
	return printSummaryWithIgnored(w, "%d of %d statefulsets are healthy.\n", s.ignored, s.healthy, s.total)
}

func (s *statefulsetsStatus) Details(w io.Writer, colored bool) error {
	return s.toTable().Fprint(w, colored)
}

func (s *statefulsetsStatus) ExitCode() int {
	if s.unhealthy > s.ignored {
		return 50
	}

	return 0
}

func (s *statefulsetsStatus) toTable() Table {
	header := []string{"Namespace", "Statefulset", "Replicas", "Ready", "Current", "Updated"}

	rows := [][]string{}
	for _, item := range s.statefulsets {
		row := []string{
			item.Namespace,
			item.Name,
			fmt.Sprintf("%d", item.Status.Replicas),
			fmt.Sprintf("%d", item.Status.ReadyReplicas),
			fmt.Sprintf("%d", item.Status.CurrentReplicas),
			fmt.Sprintf("%d", item.Status.UpdatedReplicas),
		}
		rows = append(rows, row)
	}

	return Table{
		Header: header,
		Rows:   rows,
	}
}

func (s *statefulsetsStatus) add(statefulsets []appsv1.StatefulSet) {
	s.total += len(statefulsets)

	for _, item := range statefulsets {
		if statefulsetIsHealthy(item) {
			s.healthy++
			continue
		}

		if isCiOrLabNamespace(item.Namespace) {
			s.ignored++
		}

		s.statefulsets = append(s.statefulsets, item)
		s.unhealthy++
	}
}

func statefulsetIsHealthy(item appsv1.StatefulSet) bool {
	if item.Spec.UpdateStrategy.Type == appsv1.OnDeleteStatefulSetStrategyType {
		// workaround for a bug in Kubernetes: https://github.com/kubernetes/kubernetes/issues/106055
		return item.Status.Replicas == item.Status.ReadyReplicas &&
			item.Status.Replicas == item.Status.UpdatedReplicas
	}
	return item.Status.Replicas == item.Status.ReadyReplicas &&
		item.Status.Replicas == item.Status.CurrentReplicas &&
		item.Status.Replicas == item.Status.UpdatedReplicas
}

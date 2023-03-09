package k8status

import (
	"context"
	"fmt"
	"io"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deploymentsStatus struct {
	total       int
	ignored     int
	healthy     int
	deployments []appsv1.Deployment
	unhealthy   int
}

func NewDeploymentsStatus(ctx context.Context, client *KubernetesClient) (status, error) {
	deploymentsList, err := client.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	deployments := deploymentsList.Items

	status := &deploymentsStatus{
		deployments: []appsv1.Deployment{},
	}
	status.add(deployments)

	return status, nil
}

func (s *deploymentsStatus) Summary(w io.Writer) error {
	return printSummaryWithIgnored(w, "%d of %d deployments are healthy.\n", s.ignored, s.healthy, s.total)
}

func (s *deploymentsStatus) Details(w io.Writer, colored bool) error {
	return s.toTable().Fprint(w, colored)
}

func (s *deploymentsStatus) ExitCode() int {
	if s.unhealthy > 0 {
		return 48
	}

	return 0
}

func (s *deploymentsStatus) toTable() Table {
	header := []string{"Deployment", "Namespace", "Replicas", "Available", "Up-to-date", "Ready"}

	rows := [][]string{}
	for _, item := range s.deployments {
		row := []string{
			item.Name,
			item.Namespace,
			fmt.Sprintf("%d", item.Status.Replicas),
			fmt.Sprintf("%d", item.Status.AvailableReplicas),
			fmt.Sprintf("%d", item.Status.UpdatedReplicas),
			fmt.Sprintf("%d", item.Status.ReadyReplicas),
		}
		rows = append(rows, row)
	}

	return Table{
		Header: header,
		Rows:   rows,
	}
}

func (s *deploymentsStatus) add(deployments []appsv1.Deployment) {
	s.total += len(deployments)

	for _, item := range deployments {
		if deploymentIsHealthy(item) {
			s.healthy++
			continue
		}

		if isCiOrLabNamespace(item.Namespace) {
			s.ignored++
		}

		s.deployments = append(s.deployments, item)
		s.unhealthy++
	}
}

func deploymentIsHealthy(item appsv1.Deployment) bool {
	if item.Status.Replicas == item.Status.UpdatedReplicas &&
		item.Status.Replicas == item.Status.ReadyReplicas &&
		item.Status.Replicas == item.Status.AvailableReplicas {
		return true
	}

	return false
}

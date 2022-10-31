package k8status

import (
	"context"
	"errors"
	"fmt"
	"io"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrDeploymentListIsNil error = errors.New("ErrDeploymentListIsNil")

type deploymentTableView struct {
	name      string
	namespace string
	replicas  string
	ready     string
	updated   string
	available string
}

func (c deploymentTableView) header() []string {
	return []string{"Deployment", "Namespace", "Replicas", "Available", "Up-to-date", "Ready"}
}

func (c deploymentTableView) row() []string {
	return []string{c.name, c.namespace, c.replicas, c.available, c.updated, c.ready}
}

func PrintDeploymentStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose, colored bool) (int, error) {
	deployments, err := client.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	_ = deployments
	if err != nil {
		return 0, err
	}

	return printDeploymentStatus(header, details, deployments, verbose, colored)
}

func printDeploymentStatus(header io.Writer, details io.Writer, deployments *appsv1.DeploymentList, verbose, colored bool) (int, error) {
	if deployments == nil {
		return 0, ErrDeploymentListIsNil
	}

	stats := gatherDeploymentsStats(deployments)

	err := createAndWriteDeploymentsTableInfo(header, details, stats, verbose, colored)
	if err != nil {
		return 0, err
	}

	exitCode := evaluateDeploymentsStatus(stats)

	return exitCode, nil
}

func evaluateDeploymentsStatus(stats *deploymentStats) (exitCode int) {
	exitCode = 0

	if stats.foundUnhealthyDeployment {
		return 48
	}

	return exitCode
}

func createAndWriteDeploymentsTableInfo(header io.Writer, details io.Writer, stats *deploymentStats, verbose, colored bool) error {

	table, err := CreateTable(details, deploymentTableView{}.header(), colored)
	if err != nil {
		return err
	}

	fmt.Fprintf(header, "%d of %d deployments are healthy.\n", stats.healthyDeployments, stats.deploymentsTotal)

	if verbose {
		if len(stats.tableData) != 0 {
			RenderTable(table, stats.tableData) //"renders" (not really) by writing into the details writer
		}
	}

	return nil
}

type deploymentStats struct {
	deploymentsTotal         int
	healthyDeployments       int
	tableData                [][]string
	foundUnhealthyDeployment bool
}

func gatherDeploymentsStats(deployments *appsv1.DeploymentList) *deploymentStats {
	foundUnhealthyDeployment := false

	healthy := 0
	tableData := [][]string{}

	for _, item := range deployments.Items {
		if deploymentIsHealthy(item) {
			healthy++
			continue
		}

		tv := deploymentTableView{
			item.Name,
			item.Namespace,
			fmt.Sprintf("%d", item.Status.Replicas),
			fmt.Sprintf("%d", item.Status.AvailableReplicas),
			fmt.Sprintf("%d", item.Status.UpdatedReplicas),
			fmt.Sprintf("%d", item.Status.ReadyReplicas),
		}
		tableData = append(tableData, tv.row())

		if !isCiOrLabNamespace(item.Namespace) {
			foundUnhealthyDeployment = true
		}
	}

	stats := deploymentStats{
		deploymentsTotal:         len(deployments.Items),
		healthyDeployments:       healthy,
		tableData:                tableData,
		foundUnhealthyDeployment: foundUnhealthyDeployment,
	}

	return &stats
}

func deploymentIsHealthy(item appsv1.Deployment) bool {
	if item.Status.Replicas == item.Status.UpdatedReplicas &&
		item.Status.Replicas == item.Status.ReadyReplicas &&
		item.Status.Replicas == item.Status.AvailableReplicas {
		return true
	}

	return false
}

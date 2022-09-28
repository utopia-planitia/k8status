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

var ErrNamespaceListIsNil error = errors.New("ErrNamespaceListIsNil")

type namespaceTableView struct {
	name  string
	phase string
}

func (c namespaceTableView) header() []string {
	return []string{"Namespace", "Phase"}
}

func (c namespaceTableView) row() []string {
	return []string{c.name, c.phase}
}

func PrintNamespaceStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	namespaces, err := client.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	return printNamespaceStatus(ctx, header, details, namespaces, verbose)
}

func printNamespaceStatus(_ context.Context, header io.Writer, details colorWriter, namespaces *v1.NamespaceList, verbose bool) (int, error) {
	if namespaces == nil {
		return 0, ErrNamespaceListIsNil
	}

	stats := gatherNamespacesStats(namespaces)

	err := createAndWriteNamespacesTableInfo(header, details, stats, verbose)
	if err != nil {
		return 0, err
	}

	exitCode := evaluateNamespacesStatus(stats)

	return exitCode, nil
}

func evaluateNamespacesStatus(stats *namespacesStats) (exitCode int) {
	exitCode = 0

	if stats.foundUnhealthyNamespace {
		return 43
	}

	return exitCode
}

func createAndWriteNamespacesTableInfo(header io.Writer, details colorWriter, stats *namespacesStats, verbose bool) error {

	table, err := CreateTable(details, tableHeader(namespaceTableView{}), tablewriter.FgYellowColor)
	if err != nil {
		return err
	}

	fmt.Fprintf(header, "%d of %d namespaces are healthy.\n", stats.healthyNamespaces, stats.namespacesTotal)

	if verbose {
		if len(stats.tableData) != 0 {
			RenderTable(table, stats.tableData) //"renders" (not really) by writing into the details writer
		}
	}

	return nil
}

type namespacesStats struct {
	namespacesTotal         int
	healthyNamespaces       int
	tableData               [][]string
	foundUnhealthyNamespace bool
}

func gatherNamespacesStats(namespaces *v1.NamespaceList) *namespacesStats {
	foundUnhealthyNamespace := false

	healthy := 0
	tableData := [][]string{}

	for _, item := range namespaces.Items {

		if namespaceIsHealthy(item) {
			healthy++
		} else {
			tableData = append(tableData, tableRow(namespaceTableView{item.Name, string(item.Status.Phase)}))

			foundUnhealthyNamespace = true
		}
	}

	stats := namespacesStats{
		namespacesTotal:         len(namespaces.Items),
		healthyNamespaces:       healthy,
		tableData:               tableData,
		foundUnhealthyNamespace: foundUnhealthyNamespace,
	}

	return &stats
}

func namespaceIsHealthy(item v1.Namespace) bool {
	return item.Status.Phase == v1.NamespaceActive
}

func isCiOrLabNamespace(namespace string) bool {
	return strings.HasPrefix(namespace, "ci-") ||
		strings.Contains(namespace, "-ci-") ||
		strings.HasSuffix(namespace, "-ci") ||
		strings.HasPrefix(namespace, "lab-") ||
		strings.Contains(namespace, "-lab-") ||
		strings.HasSuffix(namespace, "-lab")
}

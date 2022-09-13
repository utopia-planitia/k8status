package k8status

import (
	"context"
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintNamespaceStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	namespaces, err := client.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	healthy := 0
	table, err := CreateTable(details, []string{"Namespace", "Phase"}, tablewriter.FgGreenColor)
	if err != nil {
		return 0, err
	}
	tableData := [][]string{}

	for _, item := range namespaces.Items {
		if item.Status.Phase != v1.NamespaceActive {
			tableData = append(tableData, []string{item.Name, string(item.Status.Phase)})
			continue
		}

		healthy++
	}

	fmt.Fprintf(header, "%d of %d namespaces are active.\n", healthy, len(namespaces.Items))

	if verbose {
		if len(tableData) != 0 {
			RenderTable(table, tableData)
		}
	}

	if healthy != len(namespaces.Items) {
		return 43, nil
	}

	return 0, nil
}

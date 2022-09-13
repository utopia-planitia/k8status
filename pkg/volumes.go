package k8status

import (
	"context"
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintVolumeStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	pvs, err := client.clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	healthy := 0
	table, err := CreateTable(details, []string{"Volume", "Namespace", "Phase"}, tablewriter.FgMagentaColor)
	if err != nil {
		return 0, err
	}
	tableData := [][]string{}

	for _, item := range pvs.Items {
		if item.Status.Phase == v1.VolumeBound || item.Status.Phase == v1.VolumeAvailable {
			healthy++
		} else {
			tableData = append(tableData, []string{item.Name, item.Namespace, string(item.Status.Phase)})
		}
	}

	fmt.Fprintf(header, "%d of %d volumes are bound or available.\n", healthy, len(pvs.Items))

	if verbose {
		if len(tableData) != 0 {
			RenderTable(table, tableData)
		}
	}

	if len(pvs.Items) != healthy {
		for _, item := range pvs.Items {
			if item.Status.Phase != v1.VolumeBound && item.Status.Phase != v1.VolumeAvailable {
				fmt.Fprintf(details, "Volume %s in Namespace %s has status %s\n", item.Name, item.Namespace, item.Status.Phase)
			}
		}

		return 42, nil
	}

	return 0, nil
}

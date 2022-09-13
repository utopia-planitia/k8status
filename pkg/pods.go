package k8status

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintPodStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	pods, err := client.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	healthy := 0
	total := 0
	table, err := CreateTable(details, []string{"Pod", "Namespace", "Phase", "Containers Ready", "Containers Expected"}, tablewriter.FgRedColor)
	if err != nil {
		return 0, err
	}
	tableData := [][]string{}

	for _, item := range pods.Items {
		if item.Status.Phase == v1.PodSucceeded || item.Status.Phase == v1.PodFailed {
			continue
		}

		total++

		containerReady := 0
		for _, containerStatus := range item.Status.ContainerStatuses {
			if containerStatus.Ready {
				containerReady++
			}
		}

		if len(item.Spec.Containers) == containerReady {
			healthy++
		} else {
			tableData = append(tableData, []string{item.Name, item.Namespace, string(item.Status.Phase),
				fmt.Sprintf("%d", containerReady), fmt.Sprintf("%d", len(item.Spec.Containers))})
		}
	}

	fmt.Fprintf(header, "%d of %d pods are running.\n", healthy, total)

	if verbose {
		if len(tableData) != 0 {
			RenderTable(table, tableData)
		}
	}

	for _, item := range pods.Items {
		if item.Status.Phase == v1.PodSucceeded {
			continue
		}

		if strings.Contains(item.ObjectMeta.Namespace, "ci") || strings.Contains(item.ObjectMeta.Namespace, "lab") {
			continue
		}

		containerReady := 0
		for _, containerStatus := range item.Status.ContainerStatuses {
			if containerStatus.Ready {
				containerReady++
			}
		}

		if len(item.Spec.Containers) != containerReady {
			return 45, nil
		}
	}

	return 0, err
}

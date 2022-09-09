package k8status

import (
	"context"
	"fmt"
	"io"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintDaemonsetStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose bool) (int, error) {
	daemonsets, err := client.clientset.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	_ = daemonsets
	if err != nil {
		return 0, err
	}

	healthy := 0
	total := 0

	for _, item := range daemonsets.Items {
		total++

		if verbose {
			_, err = details.Write([]byte(fmt.Sprintf("In namespace \"%s\", daemonset \"%s\" has \"%d\""+
				" scheduled nodes but has \"%d\" current, \"%d\" ready, \"%d\" up-to-date and \"%d\"available\n",
				item.Namespace, item.Name, item.Status.DesiredNumberScheduled, item.Status.CurrentNumberScheduled,
				item.Status.NumberReady, item.Status.UpdatedNumberScheduled, item.Status.NumberAvailable)))
		}
		if err != nil {
			return 0, err
		}

		if item.Status.DesiredNumberScheduled == item.Status.CurrentNumberScheduled &&
			item.Status.DesiredNumberScheduled == item.Status.NumberReady &&
			item.Status.DesiredNumberScheduled == item.Status.UpdatedNumberScheduled &&
			item.Status.DesiredNumberScheduled == item.Status.NumberAvailable {
			healthy++
		} else {
			if verbose {
				_, err = details.Write([]byte(fmt.Sprintf("In namespace \"%s\", daemonset \"%s\" has \"%d\""+
					" scheduled nodes but has \"%d\" current, \"%d\" ready, \"%d\" up-to-date and \"%d\"available\n",
					item.Namespace, item.Name, item.Status.DesiredNumberScheduled, item.Status.CurrentNumberScheduled,
					item.Status.NumberReady, item.Status.UpdatedNumberScheduled, item.Status.NumberAvailable)))
				if err != nil {
					return 0, err
				}
			}
		}

	}

	fmt.Fprintf(header, "%d of %d daemonsets are healthy.\n", healthy, total)

	for _, item := range daemonsets.Items {

		if strings.Contains(item.Namespace, "ci") || strings.Contains(item.Namespace, "lab") {
			continue
		}

		deploymentHealthy := item.Status.DesiredNumberScheduled == item.Status.CurrentNumberScheduled &&
			item.Status.DesiredNumberScheduled == item.Status.NumberReady &&
			item.Status.DesiredNumberScheduled == item.Status.UpdatedNumberScheduled &&
			item.Status.DesiredNumberScheduled == item.Status.NumberAvailable

		if !deploymentHealthy {
			return 51, nil
		}

	}

	return 0, err

}

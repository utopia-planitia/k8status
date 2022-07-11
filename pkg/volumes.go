package k8status

import (
	"context"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintVolumeStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose bool) (int, error) {
	pvs, err := client.clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	healthy := 0
	for _, item := range pvs.Items {
		if item.Status.Phase == v1.VolumeBound || item.Status.Phase == v1.VolumeAvailable {
			healthy++
		}
	}

	fmt.Fprintf(header, "%d of %d volumes are bound or available.\n", healthy, len(pvs.Items))

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

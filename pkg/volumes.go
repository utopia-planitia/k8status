package k8status

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func PrintVolumeStatus(ctx context.Context, restconfig *rest.Config, clientset *kubernetes.Clientset, verbose bool) (int, error) {
	pvs, err := clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	healthy := 0
	for _, item := range pvs.Items {
		if item.Status.Phase == v1.VolumeBound || item.Status.Phase == v1.VolumeAvailable {
			healthy++
		}
	}
	fmt.Printf("%d of %d volumes are bound or available.\n", healthy, len(pvs.Items))

	if len(pvs.Items) != healthy {
		for _, item := range pvs.Items {
			if item.Status.Phase != v1.VolumeBound && item.Status.Phase != v1.VolumeAvailable {
				fmt.Printf("%s %s\n", item.Namespace, item.Name)
			}
		}
	}

	if healthy != len(pvs.Items) {
		return 42, nil
	}

	return 0, nil
}

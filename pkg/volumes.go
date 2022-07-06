package k8status

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func PrintVolumeStatus(ctx context.Context, clientset *kubernetes.Clientset, verbose bool) error {
	pv, err := clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	healthy := 0
	for _, item := range pv.Items {
		if item.Status.Phase == v1.VolumeBound || item.Status.Phase == v1.VolumeAvailable {
			healthy++
		}
	}
	fmt.Printf("%d of %d volumes are bound or available.\n", healthy, len(pv.Items))

	if len(pv.Items) != healthy {
		for _, item := range pv.Items {
			if item.Status.Phase != v1.VolumeBound && item.Status.Phase != v1.VolumeAvailable {
				fmt.Printf("%s %s\n", item.Namespace, item.Name)
			}
		}
	}

	return nil
}

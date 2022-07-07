package k8status

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func PrintVolumeClaimStatus(ctx context.Context, clientset *kubernetes.Clientset, verbose bool) error {
	pvcs, err := clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	healthy := 0
	for _, item := range pvcs.Items {
		if item.Status.Phase == v1.ClaimBound {
			healthy++
		}
	}
	fmt.Printf("%d of %d volume claims are bound.\n", healthy, len(pvcs.Items))

	if len(pvcs.Items) != healthy {
		for _, item := range pvcs.Items {
			if item.Status.Phase != v1.ClaimBound {
				fmt.Printf("%s %s\n", item.Namespace, item.Name)
			}
		}
	}

	return nil
}

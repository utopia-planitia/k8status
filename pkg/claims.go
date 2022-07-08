package k8status

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func PrintVolumeClaimStatus(ctx context.Context, restconfig *rest.Config, clientset *kubernetes.Clientset, verbose bool) (int, error) {
	pvcs, err := clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
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

	return 0, nil
}

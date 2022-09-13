package k8status

import (
	"context"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PrintVolumeClaimStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	pvcs, err := client.clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	healthy := 0
	for _, item := range pvcs.Items {
		if item.Status.Phase == v1.ClaimBound {
			healthy++
		}
	}

	fmt.Fprintf(header, "%d of %d volume claims are bound.\n", healthy, len(pvcs.Items))

	if len(pvcs.Items) != healthy {
		for _, item := range pvcs.Items {
			if item.Status.Phase != v1.ClaimBound {
				fmt.Fprintf(details, "Persistent Volume Claim %s in namespace %s is in phase %s\n", item.Name, item.Namespace, item.Status.Phase)
			}
		}

		return 43, nil
	}

	return 0, nil
}

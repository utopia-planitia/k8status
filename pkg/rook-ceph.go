package k8status

import (
	"bytes"
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	rookCephNamespace = "rook-ceph"
	rookCephLabel     = "app=rook-ceph-tools"
	rookCephStatusOk  = "HEALTH_OK"
)

type CephStatus struct {
	Health struct {
		Status string `json:"status"`
		Checks map[string]struct {
			Severity string `json:"severity"`
			Summary  struct {
				Message string `json:"message"`
			} `json:"summary"`
		} `json:"checks"`
	} `json:"health"`
}

func PrintRookCephStatus(ctx context.Context, restconfig *rest.Config, clientset *kubernetes.Clientset, verbose bool) (int, error) {
	exists, err := namespaceExists(ctx, clientset, rookCephNamespace)
	if err != nil {
		return 0, err
	}

	if !exists {
		if verbose {
			fmt.Printf("Rook-Ceph was not found.\n")
		}

		return 0, nil
	}

	listOptions := metav1.ListOptions{
		LabelSelector: rookCephLabel,
	}

	pods, err := listPods(ctx, clientset, rookCephNamespace, listOptions)
	if err != nil {
		if verbose {
			fmt.Printf("rook-ceph-tools was not found.\n")
		}

		return 0, nil
	}

	if len(pods) == 0 {
		return 0, fmt.Errorf("no pods found")
	}

	output := &bytes.Buffer{}
	err = exec(
		clientset,
		restconfig,
		rookCephNamespace,
		pods[0].Name,
		"ceph status --format json",
		output,
	)
	if err != nil {
		return 0, err
	}

	cephStatus := &CephStatus{}
	err = json.Unmarshal(output.Bytes(), cephStatus)
	if err != nil {
		return 0, err
	}

	if cephStatus.Health.Status == rookCephStatusOk {
		fmt.Println("Ceph is healthy.")
		return 0, nil
	}

	fmt.Println("Ceph is unhealthy.")

	if verbose {
		err = exec(
			clientset,
			restconfig,
			rookCephNamespace,
			pods[0].Name,
			"ceph status",
			os.Stdout,
		)
		if err != nil {
			return 0, err
		}
	} else {
		for _, check := range cephStatus.Health.Checks {
			fmt.Println(check.Summary.Message)
		}
	}

	// @TODO: Exit Code 48 ( status.sh, line 115	)

	if cephStatus.Health.Status != rookCephStatusOk {
		return 47, nil
	}

	return 0, nil
}

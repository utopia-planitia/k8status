package k8status

import (
	"bytes"
	"context"
	"fmt"
	"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
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

func PrintRookCephStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose bool) (int, error) {
	exists, err := namespaceExists(ctx, client.clientset, rookCephNamespace)
	if err != nil {
		return 0, err
	}

	if !exists {
		if verbose {
			fmt.Fprintf(header, "Rook-Ceph was not found.\n")
		}

		return 0, nil
	}

	listOptions := metav1.ListOptions{
		LabelSelector: rookCephLabel,
	}

	pods, err := listPods(ctx, client.clientset, rookCephNamespace, listOptions)
	if err != nil {
		if verbose {
			fmt.Fprintf(header, "rook-ceph-tools was not found.\n")
		}

		return 0, nil
	}

	if len(pods) == 0 {
		return 0, fmt.Errorf("no pods found")
	}

	output := &bytes.Buffer{}
	err = exec(
		client,
		rookCephNamespace,
		pods[0].Name,
		"",
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
		fmt.Fprintln(header, "Ceph is healthy.")
		return 0, nil
	}

	fmt.Fprintln(header, "Ceph is unhealthy.")

	if verbose {
		err = exec(
			client,
			rookCephNamespace,
			pods[0].Name,
			"",
			"ceph status",
			details,
		)
		if err != nil {
			return 0, err
		}
	} else {
		for _, check := range cephStatus.Health.Checks {
			fmt.Fprintln(details, check.Summary.Message)
		}
	}

	// @TODO: Exit Code 48 ( status.sh, line 115	)

	if cephStatus.Health.Status != rookCephStatusOk {
		return 47, nil
	}

	return 0, nil
}

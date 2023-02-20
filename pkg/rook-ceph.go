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

type rookCephStatus struct {
	found  bool
	health CephHealth
}

type CephStatus struct {
	Health CephHealth `json:"health"`
}

type CephHealth struct {
	Status string `json:"status"`
	Checks map[string]struct {
		Severity string `json:"severity"`
		Summary  struct {
			Message string `json:"message"`
		} `json:"summary"`
	} `json:"checks"`
}

func NewRookCephStatus(ctx context.Context, client *KubernetesClient) (status, error) {
	exists, err := namespaceExists(ctx, client.clientset, rookCephNamespace)
	if err != nil {
		return nil, err
	}

	status := &rookCephStatus{
		found: exists,
	}

	if !status.found {
		return status, nil
	}

	health, err := getRookCephHealth(ctx, client)
	if err != nil {
		return nil, err
	}

	status.health = health

	return status, nil
}

func getRookCephHealth(ctx context.Context, client *KubernetesClient) (CephHealth, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: rookCephLabel,
	}

	pods, err := listPods(ctx, client.clientset, rookCephNamespace, listOptions)
	if err != nil {
		return CephHealth{}, fmt.Errorf("lookup rook-ceph-tools pod: %v", err)
	}

	if len(pods) == 0 {
		return CephHealth{}, fmt.Errorf("lookup rook-ceph-tools pod: pod is missing: %v", err)
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
		return CephHealth{}, fmt.Errorf("execute ceph health check in rook-ceph pod: %v", err)
	}

	cephStatus := &CephStatus{}
	err = json.Unmarshal(output.Bytes(), cephStatus)
	if err != nil {
		return CephHealth{}, err
	}

	return cephStatus.Health, nil
}

func (s *rookCephStatus) Summary(w io.Writer) error {
	if !s.found {
		_, err := fmt.Fprintf(w, "rook-ceph was not found.\n")
		return err
	}

	status := "Ceph is healthy."
	if s.health.Status != rookCephStatusOk {
		status = "Ceph is unhealthy."
	}

	_, err := fmt.Fprintln(w, status)
	return err
}

func (s *rookCephStatus) Details(w io.Writer, colored bool) error {
	for _, check := range s.health.Checks {
		_, err := fmt.Fprintln(w, check.Summary.Message)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *rookCephStatus) ExitCode() int {
	// @TODO: Exit Code 48 ( status.sh, line 115	)

	if !s.found {
		return 0
	}

	if s.health.Status != rookCephStatusOk {
		return 47
	}

	return 0
}

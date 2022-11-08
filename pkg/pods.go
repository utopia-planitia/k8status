package k8status

import (
	"context"
	"fmt"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type podsStatus struct {
	total     int
	ignored   int
	healthy   int
	pods      []v1.Pod
	unhealthy int
}

func NewPodsStatus(ctx context.Context, client *KubernetesClient) (status, error) {
	podsList, err := client.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods := podsList.Items

	status := &podsStatus{
		pods: []v1.Pod{},
	}
	status.add(pods)

	return status, nil
}

func (s *podsStatus) Summary(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%d of %d pods are healthy (%d ignored).\n", s.healthy, s.total, s.ignored)
	return err
}

func (s *podsStatus) Details(w io.Writer, colored bool) error {
	return s.toTable().Fprint(w, colored)
}

func (s *podsStatus) ExitCode() int {
	if s.unhealthy > 0 {
		return 51
	}

	return 0
}

func (s *podsStatus) toTable() Table {
	header := []string{"Pod", "Namespace", "Phase", "Containers Ready", "Containers Expected"}

	rows := [][]string{}
	for _, item := range s.pods {
		row := []string{
			item.Name,
			item.Namespace,
			string(item.Status.Phase),
			fmt.Sprintf("%d", getReadyContainers(item)),
			fmt.Sprintf("%d", len(item.Spec.Containers)),
		}
		rows = append(rows, row)
	}

	return Table{
		Header: header,
		Rows:   rows,
	}
}

func (s *podsStatus) add(pvcs []v1.Pod) {
	s.total += len(pvcs)

	for _, item := range pvcs {
		if isCiOrLabNamespace(item.Namespace) {
			s.ignored++
			continue
		}

		if item.Status.Phase == v1.PodSucceeded || item.Status.Phase == v1.PodFailed {
			s.ignored++
			continue
		}

		if podIsHealthy(item) {
			s.healthy++
			continue
		}

		s.pods = append(s.pods, item)
		s.unhealthy++
	}
}

func podIsHealthy(item v1.Pod) bool {
	return len(item.Spec.Containers) == getReadyContainers(item)
}

func getReadyContainers(item v1.Pod) int {
	containerReady := 0
	for _, containerStatus := range item.Status.ContainerStatuses {
		if containerStatus.Ready {
			containerReady++
		}
	}
	return containerReady
}

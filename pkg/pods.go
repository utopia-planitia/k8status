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
	return printSummaryWithIgnored(w, "%d of %d pods are healthy.\n", s.ignored, s.healthy, s.total)
}

func (s *podsStatus) Details(w io.Writer, colored bool) error {
	return s.toTable().Fprint(w, colored)
}

func (s *podsStatus) ExitCode() int {
	if s.unhealthy > s.ignored {
		return 51
	}

	return 0
}

func (s *podsStatus) toTable() Table {
	header := []string{"Namespace", "Pod", "Phase", "Status", "Containers Ready", "Containers Expected", "Node"}

	rows := [][]string{}
	for _, item := range s.pods {
		status := ""
		containerStatus := []v1.ContainerStatus{}
		containerStatus = append(containerStatus, item.Status.InitContainerStatuses...)
		containerStatus = append(containerStatus, item.Status.ContainerStatuses...)
		containerStatus = append(containerStatus, item.Status.EphemeralContainerStatuses...)
		for _, s := range containerStatus {
			if s.State.Waiting != nil {
				status = s.State.Waiting.Reason
			}
			if s.State.Terminated != nil {
				status = s.State.Terminated.Reason
			}
		}
		row := []string{
			item.Namespace,
			item.Name,
			string(item.Status.Phase),
			status,
			fmt.Sprintf("%d", getReadyContainers(item)),
			fmt.Sprintf("%d", len(item.Spec.Containers)),
			item.Spec.NodeName,
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
		if podIsHealthy(item) {
			s.healthy++
			continue
		}

		ignored := isCiOrLabNamespace(item.Namespace)
		if ignored {
			s.ignored++
		}

		s.pods = append(s.pods, item)
		s.unhealthy++
	}
}

func podIsHealthy(item v1.Pod) bool {
	allReady := len(item.Spec.Containers) == getReadyContainers(item)
	succeeded := item.Status.Phase == v1.PodSucceeded
	return allReady || succeeded
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

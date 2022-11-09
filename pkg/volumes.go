package k8status

import (
	"context"
	"io"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type volumesStatus struct {
	total     int
	ignored   int
	healthy   int
	volumes   []v1.PersistentVolume
	unhealthy int
}

func NewVolumesStatus(ctx context.Context, client *KubernetesClient) (status, error) {
	volumesList, err := client.clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	volumes := volumesList.Items

	status := &volumesStatus{
		volumes: []v1.PersistentVolume{},
	}
	status.add(volumes)

	return status, nil
}

func (s *volumesStatus) Summary(w io.Writer) error {
	return printSummaryWithIgnored(w, "%d of %d volumes are bound or available.\n", s.ignored, s.healthy, s.total)
}

func (s *volumesStatus) Details(w io.Writer, colored bool) error {
	return s.toTable().Fprint(w, colored)
}

func (s *volumesStatus) ExitCode() int {
	if s.unhealthy > 0 {
		return 42
	}

	return 0
}

func (s *volumesStatus) toTable() Table {
	header := []string{"Volume", "Namespace", "Phase"}

	rows := [][]string{}
	for _, item := range s.volumes {
		row := []string{
			item.Name,
			item.Namespace,
			string(item.Status.Phase),
		}
		rows = append(rows, row)
	}

	return Table{
		Header: header,
		Rows:   rows,
	}
}

func (s *volumesStatus) add(pvcs []v1.PersistentVolume) {
	s.total += len(pvcs)

	for _, item := range pvcs {
		if isCiOrLabNamespace(item.Namespace) {
			s.ignored++
			continue
		}

		if volumeIsHealthy(item) {
			s.healthy++
			continue
		}

		s.volumes = append(s.volumes, item)
		s.unhealthy++
	}
}

func volumeIsHealthy(item v1.PersistentVolume) bool {
	return item.Status.Phase == v1.VolumeBound || item.Status.Phase == v1.VolumeAvailable
}

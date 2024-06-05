package k8status

import (
	"context"
	"fmt"
	"io"

	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type jobsStatus struct {
	total     int
	ignored   int
	healthy   int
	jobs      []v1.Job
	unhealthy int
}

func NewJobsStatus(ctx context.Context, client *KubernetesClient) (status, error) {
	jobsList, err := client.clientset.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	jobs := jobsList.Items

	status := &jobsStatus{
		jobs: []v1.Job{},
	}
	status.add(jobs)

	return status, nil
}

func (s *jobsStatus) Summary(w io.Writer) error {
	return printSummaryWithIgnored(w, "%d of %d jobs are healthy.\n", s.ignored, s.healthy, s.total)
}

func (s *jobsStatus) Details(w io.Writer, colored bool) error {
	return s.toTable().Fprint(w, colored)
}

func (s *jobsStatus) ExitCode() int {
	if s.unhealthy > s.ignored {
		return 51
	}

	return 0
}

func (s *jobsStatus) toTable() Table {
	header := []string{"Namespace", "Job", "Active", "Completions", "Succeeded", "Failed"}

	rows := [][]string{}
	for _, item := range s.jobs {
		row := []string{
			item.Namespace,
			item.Name,
			fmt.Sprintf("%d", item.Status.Active),
			fmt.Sprintf("%d", *item.Spec.Completions),
			fmt.Sprintf("%d", item.Status.Succeeded),
			fmt.Sprintf("%d", item.Status.Failed),
		}
		rows = append(rows, row)
	}

	return Table{
		Header: header,
		Rows:   rows,
	}
}

func (s *jobsStatus) add(jobs []v1.Job) {
	s.total += len(jobs)

	for _, item := range jobs {
		if jobIsHealthy(item) {
			s.healthy++
			continue
		}

		if isCiOrLabNamespace(item.Namespace) {
			s.ignored++
		}

		s.jobs = append(s.jobs, item)
		s.unhealthy++
	}
}

func jobIsHealthy(item v1.Job) bool {
	if item.Status.Active > 0 {
		return true
	}

	if *item.Spec.Completions == item.Status.Succeeded {
		return true
	}

	return false
}

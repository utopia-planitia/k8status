package k8status

import (
	"context"
	"io"
	"time"

	"github.com/aptible/supercronic/cronexpr"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type cronjobsStatus struct {
	total     int
	ignored   int
	healthy   int
	cronjobs  []batchv1.CronJob
	unhealthy int
}

func NewCronjobsStatus(ctx context.Context, client *KubernetesClient) (status, error) {
	cronjobsList, err := client.clientset.BatchV1().CronJobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	cronjobs := cronjobsList.Items

	status := &cronjobsStatus{
		cronjobs: []batchv1.CronJob{},
	}
	status.add(cronjobs)

	return status, nil
}

func (s *cronjobsStatus) Summary(w io.Writer) error {
	return printSummaryWithIgnored(w, "%d of %d cronjobs are healthy.\n", s.ignored, s.healthy, s.total)
}

func (s *cronjobsStatus) Details(w io.Writer, colored bool) error {
	return s.toTable().Fprint(w, colored)
}

func (s *cronjobsStatus) ExitCode() int {
	if s.unhealthy > s.ignored {
		return 52
	}

	return 0
}

func (s *cronjobsStatus) toTable() Table {
	header := []string{"Cronjob", "Namespace", "Status", "Last Success"}

	rows := [][]string{}
	for _, item := range s.cronjobs {
		neverSuccessful, failed100times := cronjobStatus(item)
		status := "Unknown"
		lastSucessful := ""

		if neverSuccessful {
			status = "Never successful"
			lastSucessful = ""
		} else if failed100times {
			status = "Too many missed start time (> 100)"
			lastSucessful = item.Status.LastSuccessfulTime.String()
		}

		row := []string{item.Name, item.Namespace, status, lastSucessful}
		rows = append(rows, row)
	}

	return Table{
		Header: header,
		Rows:   rows,
	}
}

func (s *cronjobsStatus) add(cronjobs []batchv1.CronJob) {
	s.total += len(cronjobs)

	for _, item := range cronjobs {
		if *item.Spec.Suspend {
			s.healthy++
			continue
		}

		neverSuccessful, failed100times := cronjobStatus(item)
		healthy := !neverSuccessful && !failed100times
		if healthy {
			s.healthy++
			continue
		}

		if isCiOrLabNamespace(item.Namespace) {
			s.ignored++
		}

		s.cronjobs = append(s.cronjobs, item)
		s.unhealthy++
	}
}

func cronjobStatus(item batchv1.CronJob) (neverSuccessful, failed100times bool) {
	if item.Status.LastSuccessfulTime == nil && item.Status.LastScheduleTime == nil {
		return false, false
	}

	if item.Status.LastSuccessfulTime == nil {
		return true, false
	}

	next100RunTimes := cronexpr.MustParse(item.Spec.Schedule).NextN(item.Status.LastSuccessfulTime.Time, 100)
	the100ScheduleTime := next100RunTimes[len(next100RunTimes)-1]
	failed100times = the100ScheduleTime.Before(time.Now())

	return false, failed100times
}

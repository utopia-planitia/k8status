package k8status

import (
	"context"
	"errors"
	"fmt"
	"io"

	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrJobListIsNil error = errors.New("ErrJobListIsNil")

type jobTableView struct {
	name        string
	namespace   string
	active      string
	completions string
	succeeded   string
	failed      string
}

func (c jobTableView) header() []string {
	return []string{"Job", "Namespace", "Active", "Completions", "Succeeded", "Failed"}
}

func (c jobTableView) row() []string {
	return []string{c.name, c.namespace, string(c.active), string(c.completions), string(c.succeeded), string(c.failed)}
}

func PrintJobStatus(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose, colored bool) (int, error) {
	jobs, err := client.clientset.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	return printJobStatus(header, details, jobs, verbose, colored)
}

func printJobStatus(header io.Writer, details io.Writer, jobs *v1.JobList, verbose, colored bool) (int, error) {
	if jobs == nil {
		return 0, ErrJobListIsNil
	}

	stats := gatherJobsStats(jobs)

	err := createAndWriteJobsTableInfo(header, details, stats, verbose, colored)
	if err != nil {
		return 0, err
	}

	exitCode := evaluateJobsStatus(stats)

	return exitCode, nil
}

func evaluateJobsStatus(stats *jobStats) (exitCode int) {
	exitCode = 0

	if stats.foundUnhealthyJob {
		return 49
	}

	return exitCode
}

func createAndWriteJobsTableInfo(header io.Writer, details io.Writer, stats *jobStats, verbose, colored bool) error {

	table, err := CreateTable(details, jobTableView{}.header(), colored)
	if err != nil {
		return err
	}

	fmt.Fprintf(header, "%d of %d jobs are healthy.\n", stats.healthyJobs, stats.jobsTotal)

	if verbose {
		if len(stats.tableData) != 0 {
			RenderTable(table, stats.tableData)
		}
	}

	return nil
}

type jobStats struct {
	jobsTotal         int
	healthyJobs       int
	tableData         [][]string
	foundUnhealthyJob bool
}

func gatherJobsStats(jobs *v1.JobList) *jobStats {
	foundUnhealthyJob := false

	healthy := 0
	tableData := [][]string{}

	for _, item := range jobs.Items {

		if jobIsHealthy(item) {
			healthy++
			continue
		}

		tv := jobTableView{
			item.Name,
			item.Namespace,
			fmt.Sprintf("%d", item.Status.Active),
			fmt.Sprintf("%d", *item.Spec.Completions),
			fmt.Sprintf("%d", item.Status.Succeeded),
			fmt.Sprintf("%d", item.Status.Failed),
		}
		tableData = append(tableData, tv.row())

		if !isCiOrLabNamespace(item.Namespace) {
			foundUnhealthyJob = true
		}
	}

	stats := jobStats{
		jobsTotal:         len(jobs.Items),
		healthyJobs:       healthy,
		tableData:         tableData,
		foundUnhealthyJob: foundUnhealthyJob,
	}

	return &stats
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

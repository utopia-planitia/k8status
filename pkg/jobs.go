package k8status

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
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

func PrintJobStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	jobs, err := client.clientset.BatchV1().Jobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	return printJobStatus(ctx, header, details, jobs, verbose)

	// healthy := 0
	// table, err := CreateTable(details, []string{"Job", "Namespace", "Active", "Completions", "Succeeded", "Failed"}, tablewriter.FgBlueColor)
	// if err != nil {
	// 	return 0, err
	// }
	// tableData := [][]string{}

	// for _, item := range jobs.Items {
	// 	if !isHealthy(item) {
	// 		tableData = append(tableData, []string{item.Name, item.Namespace,
	// 			fmt.Sprintf("%d", item.Status.Active), fmt.Sprintf("%d", *item.Spec.Completions),
	// 			fmt.Sprintf("%d", item.Status.Succeeded), fmt.Sprintf("%d", item.Status.Failed)})
	// 		continue
	// 	}

	// 	healthy++
	// }

	// fmt.Fprintf(header, "%d of %d jobs are completed.\n", healthy, len(jobs.Items))

	// if verbose {
	// 	if len(tableData) != 0 {
	// 		RenderTable(table, tableData)
	// 	}
	// }

	// for _, item := range jobs.Items {
	// 	if strings.Contains(item.ObjectMeta.Namespace, "ci") || strings.Contains(item.ObjectMeta.Namespace, "lab") {
	// 		continue
	// 	}

	// 	if isHealthy(item) {
	// 		continue
	// 	}

	// 	return 49, nil
	// }

	// return 0, nil
}

func printJobStatus(_ context.Context, header io.Writer, details colorWriter, jobs *v1.JobList, verbose bool) (int, error) {
	if jobs == nil {
		return 0, ErrJobListIsNil
	}

	stats := gatherJobsStats(jobs)

	err := createAndWriteJobsTableInfo(header, details, stats, verbose)
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

func createAndWriteJobsTableInfo(header io.Writer, details colorWriter, stats *jobStats, verbose bool) error {

	table, err := CreateTable(details, tableHeader(jobTableView{}), tablewriter.FgYellowColor)
	if err != nil {
		return err
	}

	fmt.Fprintf(header, "%d of %d jobs are healthy.\n", stats.healthyJobs, stats.jobsTotal)

	if verbose {
		if len(stats.tableData) != 0 {
			RenderTable(table, stats.tableData) //"renders" (not really) by writing into the details writer
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
		} else {
			tableData = append(tableData, tableRow(jobTableView{item.Name, item.Namespace,
				fmt.Sprintf("%d", item.Status.Active), fmt.Sprintf("%d", *item.Spec.Completions),
				fmt.Sprintf("%d", item.Status.Succeeded), fmt.Sprintf("%d", item.Status.Failed)}))

			if strings.Contains(item.Namespace, "ci") || strings.Contains(item.Namespace, "lab") {
				continue
			}
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

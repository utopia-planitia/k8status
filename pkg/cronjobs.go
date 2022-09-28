package k8status

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aptible/supercronic/cronexpr"
	"github.com/olekukonko/tablewriter"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ErrCronJobListIsNil error = errors.New("ErrCronJobListIsNil")

type cronjobTableView struct {
	name          string
	namespace     string
	status        string
	lastSucessful string
}

func (c cronjobTableView) header() []string {
	return []string{"Cronjob", "Namespace", "Status", "Last Success"}
}

func (c cronjobTableView) row() []string {
	return []string{c.name, c.namespace, c.status, c.lastSucessful}
}

func PrintCronjobStatus(ctx context.Context, header io.Writer, details colorWriter, client *KubernetesClient, verbose bool) (int, error) {
	cronjobs, err := client.clientset.BatchV1().CronJobs("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	return printCronjobStatus(ctx, header, details, cronjobs, verbose)
}

func printCronjobStatus(_ context.Context, header io.Writer, details colorWriter, cronjobs *batchv1.CronJobList, verbose bool) (int, error) {
	if cronjobs == nil {
		return 0, ErrCronJobListIsNil
	}

	stats := gatherCronjobStats(cronjobs)

	err := createAndWriteCronjobsTableInfo(header, details, stats, verbose)
	if err != nil {
		return 0, err
	}

	exitCode := evaluateCronjobsStatus(stats)

	return exitCode, nil
}

func evaluateCronjobsStatus(stats *cronjobsStats) (exitCode int) {
	exitCode = 0

	if stats.foundCronjobWithNoLastSuccessfulTime {
		return 52
	}

	if stats.foundCronjobWith100FailedRetries {
		return 53
	}

	return exitCode
}

func createAndWriteCronjobsTableInfo(header io.Writer, details colorWriter, stats *cronjobsStats, verbose bool) error {

	table, err := CreateTable(details, tableHeader(cronjobTableView{}), tablewriter.FgYellowColor)
	if err != nil {
		return err
	}

	fmt.Fprintf(header, "%d of %d cronjobs are healthy.\n", stats.healthyJobs, stats.jobsTotal)

	if verbose {
		if len(stats.tableData) != 0 {
			RenderTable(table, stats.tableData) //"renders" (not really) by writing into the details writer
		}
	}

	return nil
}

type cronjobsStats struct {
	foundCronjobWithNoLastSuccessfulTime bool
	foundCronjobWith100FailedRetries     bool
	jobsTotal                            int
	healthyJobs                          int
	tableData                            [][]string
}

func gatherCronjobStats(cronjobs *batchv1.CronJobList) *cronjobsStats {
	foundCronjobWithNoLastSuccessfulTime := false
	foundCronjobWith100FailedRetries := false
	healthy := 0
	tableData := [][]string{}

	for _, item := range cronjobs.Items {

		isSuspended, isCiLikeNamespace, hasNoSuccessfulRun, failed100Retries := cronjobStatus(item)

		if hasNoSuccessfulRun && !isSuspended && !isCiLikeNamespace {
			foundCronjobWithNoLastSuccessfulTime = true
		}

		// add job always to table logging
		if hasNoSuccessfulRun {
			tableData = append(tableData, tableRow(cronjobTableView{item.Name, item.Namespace, "Never successful", ""}))
		}

		if !hasNoSuccessfulRun && failed100Retries {
			foundCronjobWith100FailedRetries = true
			tableData = append(
				tableData,
				tableRow(cronjobTableView{
					item.Name, item.Namespace, "Too many missed start time (> 100)",
					item.Status.LastSuccessfulTime.String(),
				}),
			)
		}

		if isSuspended || isCiLikeNamespace || (!hasNoSuccessfulRun && !failed100Retries) {
			healthy++
		}
	}

	// log.Printf("DEBUG - foundCronjobWithNoLastSuccessfulTime: '%t'", foundCronjobWithNoLastSuccessfulTime)

	stats := cronjobsStats{
		foundCronjobWithNoLastSuccessfulTime: foundCronjobWithNoLastSuccessfulTime,
		foundCronjobWith100FailedRetries:     foundCronjobWith100FailedRetries,
		jobsTotal:                            len(cronjobs.Items),
		healthyJobs:                          healthy,
		tableData:                            tableData,
	}

	return &stats
}

func cronjobStatus(item batchv1.CronJob) (isSuspended, isCiLikeNamespace, hasNoSuccessfulRun, failed100Retries bool) {
	isSuspended = *item.Spec.Suspend
	isCiLikeNamespace = isCiOrLabNamespace(item.Namespace)
	hasNoSuccessfulRun = item.Status.LastSuccessfulTime == nil
	failed100Retries = false

	if !hasNoSuccessfulRun {
		next100RunTimes := cronexpr.MustParse(item.Spec.Schedule).NextN(item.Status.LastSuccessfulTime.Time, 100)
		the100ScheduleTime := next100RunTimes[len(next100RunTimes)-1]
		failed100Retries = the100ScheduleTime.Before(time.Now())
	}

	return isSuspended, isCiLikeNamespace, hasNoSuccessfulRun, failed100Retries
}

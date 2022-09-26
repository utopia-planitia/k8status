package k8status

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
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

	err := createAndWriteTableInfo(header, details, stats, verbose)
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

func isCiOrLabNamespace(namespace string) bool {
	return strings.HasPrefix(namespace, "ci-") ||
		strings.Contains(namespace, "-ci-") ||
		strings.HasSuffix(namespace, "-ci") ||
		strings.HasPrefix(namespace, "lab-") ||
		strings.Contains(namespace, "-lab-") ||
		strings.HasSuffix(namespace, "-lab")
}

func createAndWriteTableInfo(header io.Writer, details colorWriter, stats *cronjobsStats, verbose bool) error {

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

		if isCiOrLabNamespace(item.Namespace) {
			healthy++
			continue
		}

		// ignore all suspended cronjobs
		if *item.Spec.Suspend {
			healthy++
			continue
		}

		// never had successful run
		if item.Status.LastSuccessfulTime == nil {
			foundCronjobWithNoLastSuccessfulTime = true
			tableData = append(tableData, tableRow(cronjobTableView{item.Name, item.Namespace, "Never successful", ""}))
			continue
		}

		next100RunTimes := cronexpr.MustParse(item.Spec.Schedule).NextN(item.Status.LastSuccessfulTime.Time, 100)
		the100ScheduleTime := next100RunTimes[len(next100RunTimes)-1]
		the100ScheduledNextRunTimeAlreadyHappened := the100ScheduleTime.Before(time.Now())

		if the100ScheduledNextRunTimeAlreadyHappened {
			foundCronjobWith100FailedRetries = true
			tableData = append(
				tableData,
				tableRow(cronjobTableView{
					item.Name, item.Namespace, "Too many missed start time (> 100)",
					item.Status.LastSuccessfulTime.String(),
				}),
			)
			continue
		}

		healthy++
	}

	stats := cronjobsStats{
		foundCronjobWithNoLastSuccessfulTime: foundCronjobWithNoLastSuccessfulTime,
		foundCronjobWith100FailedRetries:     foundCronjobWith100FailedRetries,
		jobsTotal:                            len(cronjobs.Items),
		healthyJobs:                          healthy,
		tableData:                            tableData,
	}

	return &stats
}

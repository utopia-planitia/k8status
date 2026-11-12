package k8status

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/urfave/cli/v3"
)

type newStatus func(ctx context.Context, client *KubernetesClient) (status, error)

type status interface {
	Summary(w io.Writer) error
	Details(w io.Writer, colored bool) error
	ExitCode() int
}

type result struct {
	name     string
	summary  io.ReadWriter
	details  io.ReadWriter
	exitCode int
	err      error
}

type futures []<-chan *result

type results []*result

type check struct {
	name   string
	status newStatus
}

func Run(ctx context.Context, client *KubernetesClient, colored bool) error {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))

	checks := []check{
		{name: "NewNodeStatus", status: NewNodeStatus},
		{name: "NewCassandraStatus", status: NewCassandraStatus},
		{name: "NewRookCephStatus", status: NewRookCephStatus},
		{name: "NewVolumesStatus", status: NewVolumesStatus},
		{name: "NewVolumeClaimsStatus", status: NewVolumeClaimsStatus},
		{name: "NewNamespacesStatus", status: NewNamespacesStatus},
		{name: "NewDaemonsetsStatus", status: NewDaemonsetsStatus},
		{name: "NewStatefulsetsStatus", status: NewStatefulsetsStatus},
		{name: "NewDeploymentsStatus", status: NewDeploymentsStatus},
		{name: "NewCronjobsStatus", status: NewCronjobsStatus},
		{name: "NewJobsStatus", status: NewJobsStatus},
		{name: "NewPodsStatus", status: NewPodsStatus},
	}

	futures := futures{}

	for _, check := range checks {
		future := make(chan *result)
		futures = append(futures, future)

		go func(future chan *result, name string, newCheck newStatus) {
			result := &result{
				name: name,
			}

			check, err := newCheck(ctx, client)
			if err != nil {
				result.err = err
				result.exitCode = 1
				future <- result
				return
			}

			result.exitCode = check.ExitCode()

			result.summary = &bytes.Buffer{}
			err = check.Summary(result.summary)
			if err != nil {
				result.err = err
				future <- result
				return
			}

			result.details = &bytes.Buffer{}
			err = check.Details(result.details, colored)
			if err != nil {
				result.err = err
				future <- result
				return
			}

			future <- result
		}(future, check.name, check.status)
	}

	results := futures.Await()

	err := results.Errors(os.Stdout)
	if err != nil {
		return err
	}

	err = results.Summaries(os.Stdout)
	if err != nil {
		return err
	}

	err = results.Details(os.Stdout)
	if err != nil {
		return err
	}

	exitCode := results.ExitCode()
	if exitCode != 0 {
		fmt.Println()
		return cli.Exit("an issue was found", exitCode)
	}

	return nil
}

func (futures futures) Await() results {
	results := results{}

	for _, future := range futures {
		results = append(results, <-future)
	}

	return results
}

func (results results) Errors(w io.Writer) error {
	for _, result := range results {
		if result.err == nil {
			continue
		}

		_, err := fmt.Fprintf(w, "%v error: %v\n", result.name, result.err)
		if err != nil {
			return err
		}
	}

	return nil
}

func (results results) Summaries(w io.Writer) error {
	for _, result := range results {
		if result.summary == nil {
			continue
		}

		_, err := io.Copy(w, result.summary)
		if err != nil {
			return err
		}
	}

	return nil
}

func (results results) Details(w io.Writer) error {
	for _, result := range results {
		if result.details == nil {
			continue
		}

		_, err := io.Copy(w, result.details)
		if err != nil {
			return err
		}
	}

	return nil
}

func (results results) ExitCode() int {
	for _, result := range results {
		if result.exitCode != 0 {
			return result.exitCode
		}
	}

	return 0
}

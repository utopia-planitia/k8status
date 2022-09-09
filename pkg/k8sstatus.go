package k8status

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

type statusCheck func(ctx context.Context, header io.Writer, details io.Writer, client *KubernetesClient, verbose bool) (int, error)

type result struct {
	head     io.ReadWriter
	details  io.ReadWriter
	exitCode int
	err      error
}

type futures []<-chan result

type results []result

func Run(ctx context.Context, client *KubernetesClient, verbose bool) error {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))

	checks := []statusCheck{
		PrintNodeStatus,
		PrintCassandraStatus,
		PrintRookCephStatus,
		PrintVolumeStatus,
		PrintNamespaceStatus,
		PrintVolumeClaimStatus,
		PrintPodStatus,
		PrintJobStatus,
		PrintDeploymentStatus,
		PrintStatefulsetStatus,
		PrintDaemonsetStatus,
	}

	futures := futures{}

	for _, check := range checks {
		future := make(chan result)
		futures = append(futures, future)

		go func(future chan result, check statusCheck) {
			head := &bytes.Buffer{}
			details := &bytes.Buffer{}
			exitCode, err := check(ctx, head, details, client, verbose)
			future <- result{
				head:     head,
				details:  details,
				exitCode: exitCode,
				err:      err,
			}
		}(future, check)
	}

	results := futures.Await()

	err := results.Headers(os.Stdout)
	if err != nil {
		return err
	}

	fmt.Println()

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
	results := []result{}

	for _, future := range futures {
		results = append(results, <-future)
	}

	return results
}

func (results results) Headers(w io.Writer) error {
	for _, result := range results {
		_, err := io.Copy(w, result.head)
		if err != nil {
			return err
		}
	}

	return nil
}

func (results results) Details(w io.Writer) error {
	for _, result := range results {
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

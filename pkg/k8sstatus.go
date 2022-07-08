package k8status

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Check func(ctx context.Context, restconfig *rest.Config, clientset *kubernetes.Clientset, verbose bool) (int, error)

func Run(ctx context.Context, restconfig *rest.Config, clientset *kubernetes.Clientset, verbose bool) error {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))

	checks := []Check{
		PrintNodeStatus,
		PrintCassandraStatus,
		PrintRookCephStatus,
		PrintVolumeStatus,
		PrintNamespaceStatus,
		PrintVolumeClaimStatus,
		PrintPodStatus,
		PrintJobStatus,
	}

	for _, check := range checks {
		exitCode, err := check(ctx, restconfig, clientset, verbose)
		if err != nil {
			return err
		}

		if exitCode != 0 {
			return cli.NewExitError("an issue was found", exitCode)
		}
	}

	return nil
}

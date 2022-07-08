package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	cli "github.com/urfave/cli/v2"
	k8status "gitlab.com/utopia-planitia/k8status/pkg"
)

var (
	gitHash string
	gitRef  string
	verbose = &cli.BoolFlag{
		Name:  "verbose",
		Value: false,
		Usage: "Print verbose outputs.",
	}
	kubeConfigFile = &cli.StringFlag{
		Name:    "kubeconfig",
		Value:   "", // overwritten by init function
		Usage:   "Print verbose outputs.",
		EnvVars: []string{"KUBECONFIG"},
	}
	app = &cli.App{
		Name:   "K8status",
		Usage:  "A quick overview about the health of a Kubernets cluster and its workloads.",
		Action: run,
		Flags: []cli.Flag{
			verbose,
			kubeConfigFile,
		},
		Commands: []*cli.Command{
			{
				Name:   "run",
				Usage:  "Show the health overview.",
				Action: run,
			},
			{
				Name:   "version",
				Usage:  "Print the version.",
				Action: version,
			},
		},
	}
)

func init() {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("look up home directory: %v", err)
	}

	kubeConfigFile.Value = filepath.Join(dir, ".kube", "config")
}

func main() {
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	verbose := c.Bool(verbose.Name)
	kubeConfigFile := c.String(kubeConfigFile.Name)
	ctx := c.Context

	restconfig, clientset, err := k8status.KubernetesClient(kubeConfigFile)
	if err != nil {
		return err
	}

	return k8status.Run(ctx, restconfig, clientset, verbose)
}

func version(c *cli.Context) error {
	_, err := fmt.Printf("version: %s\ngit commit: %s\n", gitRef, gitHash)
	if err != nil {
		return err
	}

	return nil
}

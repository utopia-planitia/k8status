package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	supportscolor "github.com/jwalton/go-supportscolor"
	cli "github.com/urfave/cli/v2"
	k8status "gitlab.com/utopia-planitia/k8status/pkg"
)

var (
	commit         string
	version        string
	date           string
	kubeConfigFile = &cli.StringFlag{
		Name:    "kubeconfig",
		Value:   "", // overwritten by init function
		Usage:   "Path to kube config file.",
		EnvVars: []string{"KUBECONFIG"},
	}
	app = &cli.App{
		Name:   "K8status",
		Usage:  "A quick overview about the health of a Kubernets cluster and its workloads.",
		Action: run,
		Flags: []cli.Flag{
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
				Action: printVersion,
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
	kubeConfigFile := c.String(kubeConfigFile.Name)
	ctx := c.Context
	colored := supportscolor.Stdout().SupportsColor

	k8sClient, err := k8status.NewKubernetesClient(kubeConfigFile)
	if err != nil {
		return err
	}

	return k8status.Run(ctx, k8sClient, colored)
}

func printVersion(c *cli.Context) error {
	_, err := fmt.Printf("version: %s\ngit commit: %s\ngit commit date: %s\n", version, commit, date)
	if err != nil {
		return err
	}

	return nil
}

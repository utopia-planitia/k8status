package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	cli "github.com/urfave/cli/v2"
	k8status "gitlab.com/utopia-planitia/k8status/pkg"
)

var (
	gitHash string
	gitRef  string
	homeDir string
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

	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))

	restconfig, clientset, err := k8status.KubernetesClient(kubeConfigFile)
	if err != nil {
		return err
	}

	err = k8status.PrintNodeStatus(ctx, clientset, verbose)
	if err != nil {
		return err
	}

	err = k8status.PrintCassandraStatus(ctx, clientset, verbose)
	if err != nil {
		return err
	}

	_ = restconfig

	/*
		corev1 := clientset.CoreV1()
		foo, _ := corev1.Secrets("cassandra").List(ctx, metav1.ListOptions{})
		baz, _ := corev1.Secrets("cassandra").Get(ctx, "k8ssandra-superuser", metav1.GetOptions{})
		println(foo)
		println(baz)

	*/
	return nil
}

func version(c *cli.Context) error {
	_, err := fmt.Printf("version: %s\ngit commit: %s\n", gitRef, gitHash)
	if err != nil {
		return err
	}

	return nil
}

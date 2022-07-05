package main

import (
	"context"
	"fmt"
	k8status "gostatus/pkg"
	"log"
	"time"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"))

	clientset, err := k8status.GetClientSet()
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = k8status.PrintNodeStatus(ctx, clientset)
	if err != nil {
		return err
	}

	err = k8status.PrintCassandraStatus(ctx, clientset)
	if err != nil {
		return err
	}

	/*
		corev1 := clientset.CoreV1()
		foo, _ := corev1.Secrets("cassandra").List(ctx, metav1.ListOptions{})
		baz, _ := corev1.Secrets("cassandra").Get(ctx, "k8ssandra-superuser", metav1.GetOptions{})
		println(foo)
		println(baz)

	*/
	return nil
}

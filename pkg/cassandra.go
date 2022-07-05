package k8status

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func PrintCassandraStatus(ctx context.Context, clientset *kubernetes.Clientset) error {
	corev1 := clientset.CoreV1()
	cassandraNs, err := corev1.Namespaces().Get(ctx, "cassandras", metav1.GetOptions{})
	println(cassandraNs)
	println(err)

	return err
}

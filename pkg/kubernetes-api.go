package k8status

import (
	"context"
	"io"
	"net/http"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

func namespaceExists(ctx context.Context, clientset *kubernetes.Clientset, namespace string) (bool, error) {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})

	if err == nil {
		return true, nil
	}

	if errors.IsNotFound(err) {
		return false, nil
	}

	return false, err
}

func listPods(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	namespace string,
	listOptions metav1.ListOptions,
) ([]v1.Pod, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func exec(
	client *KubernetesClient,
	namespace string,
	pod string,
	container string,
	command string,
	stdout io.Writer,
) error {
	request := client.clientset.
		CoreV1().
		RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(pod).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Command:   []string{"/bin/sh", "-c", command},
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
			Container: container,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(client.restconfig, http.MethodPost, request.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return err
	}

	return nil
}

package k8status

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

type KubeClient struct {
	clientset  *kubernetes.Clientset
	restconfig *rest.Config
}

func GetClientSet() (*kubernetes.Clientset, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	kubeconfigFile := filepath.Join(home, ".kube", "config")

	kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func GetKubeClient() (*KubeClient, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	kubeconfigFile := filepath.Join(home, ".kube", "config")

	kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	client := &KubeClient{
		clientset:  clientset,
		restconfig: kubeconfig,
	}

	return client, nil
}

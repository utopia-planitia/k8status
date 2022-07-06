package k8status

import (
	"errors"
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	inClusterSentinelFile = "/run/secrets/kubernetes.io/serviceaccount/namespace"
)

func KubernetesClient(kubeconfigFile string) (*rest.Config, *kubernetes.Clientset, error) {
	restconfig, err := restConfig(kubeconfigFile)
	if err != nil {
		return nil, nil, fmt.Errorf("load kubernetes client config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("setup kubernetes client: %v", err)
	}

	return restconfig, clientset, nil
}

func restConfig(kubeConfigFile string) (*rest.Config, error) {
	inCluster, err := hasInClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("look up in cluster config: %v", err)
	}

	if inCluster {
		return inClusterConfig()
	}

	return localKubeConfig(kubeConfigFile)
}

func hasInClusterConfig() (bool, error) {
	_, err := os.Stat(inClusterSentinelFile)

	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

func inClusterConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("load in cluster config: %v", err)
	}

	return config, nil
}

func localKubeConfig(kubeConfigFile string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigFile)
	if err != nil {
		return nil, fmt.Errorf("load local kube config: %v", err)
	}

	return config, nil
}

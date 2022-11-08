package k8status

import (
	"context"
	"fmt"
	"io"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type namespacesStatus struct {
	total      int
	ignored    int
	healthy    int
	namespaces []v1.Namespace
	unhealthy  int
}

func NewNamespacesStatus(ctx context.Context, client *KubernetesClient) (status, error) {
	namespacesList, err := client.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	namespaces := namespacesList.Items

	status := &namespacesStatus{
		namespaces: []v1.Namespace{},
	}
	status.add(namespaces)

	return status, nil
}

func (s *namespacesStatus) Summary(w io.Writer, verbose bool) error {
	_, err := fmt.Fprintf(w, "%d of %d namespaces are healthy (%d ignored).\n", s.healthy, s.total, s.ignored)
	return err
}

func (s *namespacesStatus) Details(w io.Writer, verbose, colored bool) error {
	return s.toTable().Fprint(w, colored)
}

func (s *namespacesStatus) ExitCode() int {
	if s.unhealthy > 0 {
		return 43
	}

	return 0
}

func (s *namespacesStatus) toTable() Table {
	header := []string{"Namespace", "Phase"}

	rows := [][]string{}
	for _, item := range s.namespaces {
		row := []string{
			item.Name,
			string(item.Status.Phase),
		}
		rows = append(rows, row)
	}

	return Table{
		Header: header,
		Rows:   rows,
	}
}

func (s *namespacesStatus) add(namespaces []v1.Namespace) {
	s.total += len(namespaces)

	for _, item := range namespaces {
		if isCiOrLabNamespace(item.Namespace) {
			s.ignored++
			continue
		}

		if namespaceIsHealthy(item) {
			s.healthy++
			continue
		}

		s.namespaces = append(s.namespaces, item)
		s.unhealthy++
	}
}

func namespaceIsHealthy(item v1.Namespace) bool {
	return item.Status.Phase == v1.NamespaceActive
}

func isCiOrLabNamespace(namespace string) bool {
	return strings.HasPrefix(namespace, "ci-") ||
		strings.Contains(namespace, "-ci-") ||
		strings.HasSuffix(namespace, "-ci") ||
		strings.HasPrefix(namespace, "lab-") ||
		strings.Contains(namespace, "-lab-") ||
		strings.HasSuffix(namespace, "-lab")
}

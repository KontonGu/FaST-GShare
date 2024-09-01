package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s.io/klog"
)

func MakeNamespacesLister(defaultNamespace string, clusterRole bool, clientset kubernetes.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Body != nil {
			defer r.Body.Close()
		}

		namesapces := []string{}
		if clusterRole {
			namesapces = ListNamespaces(defaultNamespace, clientset)
		} else {
			namesapces = append(namesapces, defaultNamespace)
		}

		out, err := json.Marshal(namesapces)
		if err != nil {
			klog.Errorf("Failed to list namespaces: %s", err.Error())
			http.Error(w, "Failed to list namespaces", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(out)

	}
}

type NamespaceResolver func(r *http.Request) (namspace string, err error)

func NewNamsespaceResolver(defaultNamespace string, kube kubernetes.Interface) NamespaceResolver {
	return func(r *http.Request) (string, error) {
		req := struct{ Namespace string }{Namespace: defaultNamespace}

		switch r.Method {
		case http.MethodGet:
			q := r.URL.Query()
			namespace := q.Get("namespace")

			if len(namespace) > 0 {
				req.Namespace = namespace
			}
		case http.MethodPost, http.MethodPut, http.MethodDelete:
			body, _ := io.ReadAll(r.Body)
			err := json.Unmarshal(body, &req)
			if err != nil {
				log.Printf("error while getting namsespace: %s\n", err)
				return "", fmt.Errorf("unable to unmarshal json request")
			}

			r.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		allowedNamespaces := ListNamespaces(defaultNamespace, kube)
		ok := findNamespace(req.Namespace, allowedNamespaces)
		if !ok {
			return req.Namespace, fmt.Errorf("unable to manage secretes within the %s namespace", req.Namespace)
		}

		return req.Namespace, nil
	}
}

func ListNamespaces(defaultNamespace string, clientset kubernetes.Interface) []string {
	listOptions := metav1.ListOptions{}

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), listOptions)

	set := []string{}

	if err != nil {
		log.Printf("Error listing namespaces: %s", err.Error())
		set = append(set, defaultNamespace)
		return set
	}

	for _, n := range namespaces.Items {
		if _, ok := n.Annotations["fast-gshare"]; ok {
			set = append(set, n.Name)
		}
	}

	if !findNamespace(defaultNamespace, set) {
		set = append(set, defaultNamespace)
	}

	return set

}

func findNamespace(target string, items []string) bool {
	for _, n := range items {
		if n == target {
			return true
		}
	}
	return false
}

package shalm

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type k8sClient struct {
	client *rest.RESTClient
}

type request struct {
	request   *rest.Request
	namespace *string
	resource  string
	name      string
}

type result struct {
	rest.Result
	err error
}

var errUnknownResource = errors.New("Unknown resource")

var kindToGroupVersionKind = map[string]schema.GroupVersionKind{}

func init() {
	for k := range scheme.Scheme.AllKnownTypes() {
		groupVersion := schema.GroupVersionKind{Version: k.Version, Group: k.Group, Kind: strings.ToLower(k.Kind) + "s"}
		if len(groupVersion.Group) == 0 {
			groupVersion.Group = "api"
		}
		kindToGroupVersionKind[strings.ToLower(k.Kind)] = groupVersion
		kindToGroupVersionKind[groupVersion.Kind] = groupVersion
	}
}

func configKube(kubeConfig string) (*rest.Config, error) {
	if len(kubeConfig) == 0 {
		host := os.Getenv("KUBERNETES_SERVICE_HOST")
		if len(host) != 0 {
			return rest.InClusterConfig()
		} else {
			env, ok := os.LookupEnv("KUBECONFIG")
			if ok {
				kubeConfig = env
			} else {
				path := filepath.Join(homeDir(), ".kube", "config")
				kubeConfig = path
			}
		}
	}
	return clientcmd.BuildConfigFromFlags("", kubeConfig)

}
func newK8sClient(config *rest.Config) (*k8sClient, error) {
	config.NegotiatedSerializer = scheme.Codecs
	client, err := rest.UnversionedRESTClientFor(config)
	if err != nil {
		return nil, err
	}
	scheme.Scheme.AllKnownTypes()
	return &k8sClient{client: client}, nil

}

func (k *k8sClient) Get() request {
	return request{request: k.client.Get()}
}

func (r request) Namespace(namespace *string) request {
	r.namespace = namespace
	return r
}

func (r request) Resource(resource string) request {
	r.resource = resource
	return r
}

func (r request) Name(name string) request {
	r.name = name
	return r
}

func (r request) Do() result {
	gv, ok := kindToGroupVersionKind[strings.ToLower(r.resource)]
	if !ok {
		return result{err: errUnknownResource}
	}
	if r.namespace != nil {
		r.request.AbsPath(gv.Group, gv.Version, "namespaces", *r.namespace, gv.Kind, r.name)
	} else {
		r.request.AbsPath(gv.Group, gv.Version, gv.Kind, r.name)
	}
	return result{Result: r.request.Do()}
}

func (r result) Get() (*Object, error) {
	if r.err != nil {
		return nil, r.err
	}
	data, err := r.Result.Raw()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("0-length response")
	}
	var obj Object
	if err = json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	if obj.Kind == "Status" {
		return nil, errors.New(string(data))
	}
	return &obj, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

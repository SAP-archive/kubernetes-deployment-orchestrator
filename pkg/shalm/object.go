package shalm

import (
	"encoding/json"
	"strings"
)

// MetaData -
type MetaData struct {
	Namespace   string
	Name        string
	Labels      map[string]string
	Annotations map[string]string
	Additional  map[string]json.RawMessage
}

// Object -
type Object struct {
	MetaData   MetaData                   `json:"metadata,omitempty"`
	APIVersion string                     `json:"apiVersion,omitempty"`
	Kind       string                     `json:"kind,omitempty"`
	Additional map[string]json.RawMessage `json:",inline"`
}

// MarshalJSON -
func (m MetaData) MarshalJSON() ([]byte, error) {
	r := map[string]json.RawMessage{}
	for k, v := range m.Additional {
		r[k] = v
	}
	add(r, "namespace", m.Namespace)
	add(r, "name", m.Name)
	if len(m.Labels) != 0 {
		add(r, "labels", m.Labels)
	}
	if len(m.Annotations) != 0 {
		add(r, "annotations", m.Annotations)
	}
	return json.Marshal(r)
}

// UnmarshalJSON -
func (m *MetaData) UnmarshalJSON(b []byte) error {
	var r map[string]json.RawMessage
	if err := json.Unmarshal(b, &r); err != nil {
		return err
	}
	remove(r, "namespace", &m.Namespace)
	remove(r, "name", &m.Name)
	remove(r, "labels", &m.Labels)
	remove(r, "annotations", &m.Annotations)
	m.Additional = r
	return nil
}

// MarshalJSON -
func (o Object) MarshalJSON() ([]byte, error) {
	r := map[string]json.RawMessage{}
	for k, v := range o.Additional {
		r[k] = v
	}
	add(r, "metadata", o.MetaData)
	add(r, "apiVersion", o.APIVersion)
	add(r, "kind", o.Kind)
	return json.Marshal(r)
}

// UnmarshalJSON -
func (o *Object) UnmarshalJSON(b []byte) error {
	var r map[string]json.RawMessage
	if err := json.Unmarshal(b, &r); err != nil {
		return err
	}
	remove(r, "metadata", &o.MetaData)
	remove(r, "apiVersion", &o.APIVersion)
	remove(r, "kind", &o.Kind)
	o.Additional = r
	return nil
}

func (o *Object) setDefaultNamespace(namespace string) {
	if isNameSpaced(o.Kind) && o.MetaData.Namespace == "" {
		o.MetaData.Namespace = namespace
	}
}

func (o *Object) kindOrdinal() int {
	switch o.Kind {
	case "Namespace":
		return 1
	case "NetworkPolicy":
		return 2
	case "ResourceQuota":
		return 3
	case "LimitRange":
		return 4
	case "PodSecurityPolicy":
		return 5
	case "PodDisruptionBudget":
		return 6
	case "Secret":
		return 7
	case "ConfigMap":
		return 8
	case "StorageClass":
		return 9
	case "PersistentVolume":
		return 10
	case "PersistentVolumeClaim":
		return 11
	case "ServiceAccount":
		return 12
	case "CustomResourceDefinition":
		return 13
	case "ClusterRole":
		return 14
	case "ClusterRoleList":
		return 15
	case "ClusterRoleBinding":
		return 16
	case "ClusterRoleBindingList":
		return 17
	case "Role":
		return 18
	case "RoleList":
		return 19
	case "RoleBinding":
		return 20
	case "RoleBindingList":
		return 21
	case "Service":
		return 22
	case "DaemonSet":
		return 23
	case "Pod":
		return 24
	case "ReplicationController":
		return 25
	case "ReplicaSet":
		return 26
	case "Deployment":
		return 27
	case "HorizontalPodAutoscaler":
		return 28
	case "StatefulSet":
		return 29
	case "Job":
		return 30
	case "CronJob":
		return 31
	case "Ingress":
		return 32
	case "APIService":
		return 33
	default:
		return 1000
	}
}

func add(m map[string]json.RawMessage, key string, v interface{}) {
	b, _ := json.Marshal(v)
	if len(b) <= 2 {
		return
	}
	m[key] = json.RawMessage(b)
}

func remove(m map[string]json.RawMessage, key string, v interface{}) {
	r, ok := m[key]
	if !ok {
		return
	}
	_ = json.Unmarshal(r, v)
	delete(m, key)
}

func isNameSpaced(kind string) bool {
	switch strings.ToLower(kind) {
	case "namespace":
		return false
	case "resourcequota":
		return false
	case "customresourcedefinition":
		return false
	case "clusterrole":
		return false
	case "clusterrolelist":
		return false
	case "clusterrolebinding":
		return false
	case "clusterrolebindinglist":
		return false
	case "apiservice":
		return false
	}
	return true
}

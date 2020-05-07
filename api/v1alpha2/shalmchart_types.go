/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	"encoding/json"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/util/intstr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ChartSpec defines the desired state of ShalmChart
type ChartSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	// Values that should be merged in the chart on the installation side
	Values runtime.RawExtension `json:"values,omitempty"`
	// +optional
	// Args which are passed to the constructor of the chart
	Args []intstr.IntOrString `json:"args,omitempty"`
	// +optional
	// KwArgs which are passed to the constructor of the chart
	KwArgs runtime.RawExtension `json:"kwargs,omitempty"`
	// +optional
	// Kubeconfig which is used for the installation. Install into local cluster if empty
	KubeConfig string `json:"kubeconfig,omitempty"`
	// Namespace which is used for the installation
	Namespace string `json:"namespace,omitempty"`
	// +optional
	// Suffix which is used to make the chart instance unique
	Suffix string `json:"suffix,omitempty"`
	// +optional
	// ChartTgz containts the complete chart
	ChartTgz []byte `json:"chart_tgz,omitempty"`
	// +optional
	// ChartURL containts the URL for the chart. If empty the ChartTgz field is used.
	ChartURL string `json:"chart_url,omitempty"`
	// +optional
	// Tool which is used to do the deployment and deletion
	Tool string `json:"tool,omitempty"`
}

// SetKwArgs set KwArgs member
func (s *ChartSpec) SetKwArgs(kwargs map[string]interface{}) error {
	rawKwArgs, err := json.Marshal(kwargs)
	if err != nil {
		return errors.Wrapf(err, "error during marshalling of kwargs")
	}
	s.KwArgs.Raw = rawKwArgs
	return nil
}

// GetKwArgs get KwArgs member
func (s *ChartSpec) GetKwArgs() (map[string]interface{}, error) {
	if len(s.KwArgs.Raw) > 0 {
		var result map[string]interface{}
		err := json.Unmarshal(s.KwArgs.Raw, &result)
		if err != nil {
			return nil, errors.Wrapf(err, "error during unmarshalling of kwargs")
		}
		return result, nil
	}
	return nil, nil
}

// SetValues set values member
func (s *ChartSpec) SetValues(values map[string]interface{}) error {
	rawValues, err := json.Marshal(values)
	if err != nil {
		return errors.Wrapf(err, "error during marshalling of values")
	}
	s.Values.Raw = rawValues
	return nil
}

// GetValues get values member
func (s *ChartSpec) GetValues() (map[string]interface{}, error) {
	if len(s.Values.Raw) > 0 {
		var result map[string]interface{}
		err := json.Unmarshal(s.Values.Raw, &result)
		if err != nil {
			return nil, errors.Wrapf(err, "error during unmarshalling of values")
		}
		return result, nil
	}
	return nil, nil
}

// Operation defines the progress of the last operation
type Operation struct {
	// Type of installation can be apply or delete
	Type string `json:"type"`
	// Progress of installation in percent
	Progress int `json:"progress"`
}

// ChartStatus defines the observed state of ShalmChart
type ChartStatus struct {

	// LastOp containts the last operation status
	// +optional
	LastOp Operation `json:"lastOp,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.status.lastOp.type`
// +kubebuilder:printcolumn:name="Progress",type=integer,JSONPath=`.status.lastOp.progress`

// ShalmChart is the Schema for the shalmcharts API
type ShalmChart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ChartSpec `json:"spec,omitempty"`
	// +optional
	Status ChartStatus `json:"status,omitempty"`
}

// Init -
func (s *ShalmChart) Init() {
	if s.Spec.Values.Raw == nil {
		s.Spec.Values.Raw = []byte("{}")
	}
	if s.Spec.KwArgs.Raw == nil {
		s.Spec.KwArgs.Raw = []byte("{}")
	}
}

// +kubebuilder:object:root=true

// ShalmChartList contains a list of ShalmChart
type ShalmChartList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ShalmChart `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ShalmChart{}, &ShalmChartList{})
}

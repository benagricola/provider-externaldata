/*
Copyright 2021 The Crossplane Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtime "k8s.io/apimachinery/pkg/runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// SourceType is the type of external data source to retrieve
// values from.
// +kubebuilder:validation:Enum=configmap;url
type SourceType string

// SourceTypeConfigMap is a Config Map Source
const SourceTypeConfigMap SourceType = "configmap"

// SourceTypeURL is a URL
const SourceTypeURL SourceType = "url"

// DataSourceParameters are the configurable fields of a DataSource.
type DataSourceParameters struct {
	SourceType SourceType `json:"type"`

	// ConfigMapName is the name of a Kubernetes ConfigMap to look up
	// in the Namespace configured on the current ProviderConfig, when
	// type is 'configmap'
	// +optional
	ConfigMapName *string `json:"configMapName,omitempty"`

	// URL is the URL of a JSON endpint to retrieve data from, when type
	// is 'url'
	// +optional
	URL *string `json:"url,omitempty"`
}

// A DataSourceSpec defines the desired state of a DataSource.
type DataSourceSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DataSourceParameters `json:"forProvider"`
}

// A DataSourceStatus represents the observed state of a DataSource.
type DataSourceStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider contains the results of our external data lookup.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	AtProvider *runtime.RawExtension `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A DataSource retrieves data from an external source such as a
// Kubernetes ConfigMap or HTTP endpoint that returns JSON.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".spec.forProvider.type"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,externaldata}
type DataSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataSourceSpec   `json:"spec"`
	Status DataSourceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DataSourceList contains a list of DataSource
type DataSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataSource `json:"items"`
}

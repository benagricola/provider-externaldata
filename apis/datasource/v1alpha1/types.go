/*
Copyright 2020 The Crossplane Authors.

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

type SourceType string

const SourceTypeConfigMap SourceType = "configmap"

// DataSourceParameters are the configurable fields of a DataSource.
type DataSourceParameters struct {
	// +kubebuilder:validation:Enum=configmap
	SourceType SourceType `json:"type"`

	// +optional
	ConfigMapRef xpv1.Reference `json:"configMapRef,omitempty"`

	// +optional
	ConfigMapSelector xpv1.Selector `json:"configMapSelector,omitempty"`
}

// A DataSourceSpec defines the desired state of a DataSource.
type DataSourceSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DataSourceParameters `json:"forProvider"`
}

// A DataSourceStatus represents the observed state of a DataSource.
type DataSourceStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	AtProvider          runtime.RawExtension `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A DataSource is an example API type.
// +kubebuilder:subresource:status
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

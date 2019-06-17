package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FedoraSpec defines the desired state of Fedora
// +k8s:openapi-gen=true
type FedoraSpec struct {
	OSVersion string `json:"osVersion,omitempty"`
	VMName    string `json:"vmName,omitempty"`
	Memory    string `json:"memory,omitempty"`
	CPUCores  uint32 `json:"cpuCores,omitempty"`
	CloudInit string `json:"cloudInit,omitempty"`
}

// FedoraStatus defines the observed state of Fedora
// +k8s:openapi-gen=true
type FedoraStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	VMs []string `json:"vms"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Fedora is the Schema for the fedoras API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Fedora struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FedoraSpec   `json:"spec,omitempty"`
	Status FedoraStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FedoraList contains a list of Fedora
type FedoraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Fedora `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Fedora{}, &FedoraList{})
}

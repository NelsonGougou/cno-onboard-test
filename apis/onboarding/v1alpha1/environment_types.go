package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type User struct {
	ID            string   	`json:"id"`
	Username      string 	`json:"username" validate:"required"`
	Email         string 	`json:"email"`
	UserFullName  string 	`json:"userFullName"`
	EnvironmentID string 	`json:"environmentId"`
	Role          string    `json:"role" validate:"required"`
}

// EnvironmentSpec defines the desired state of Environment
type EnvironmentSpec struct {
	Name      string `json:"name" validate:"required"`
	IsProd	  bool   `json:"isprod"`
	Resources `json:"resources" validate:"required"`
	// +kubebuilder:validation:Pattern=`^(\d+(e\d+)?|\d+(\.\d+)?(e\d+)?[EPTGMK]i?)$`
	Storage    string   `json:"storage" validate:"required"`
	Users      []User   `json:"users"`
}

// EnvironmentStatus defines the observed state of Environment (success, failed)
type EnvironmentStatus struct {
	EnvironmentStatus string `json:"environmentStatus"`
}

// Environment is the Schema for the environments API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=environments,scope=Cluster
// +kubebuilder:storageversion
type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvironmentSpec   `json:"spec,omitempty"`
	Status EnvironmentStatus `json:"status,omitempty"`
}

// EnvironmentList contains a list of Environment
type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Environment `json:"items"`
}

// ResourceDescription describes CPU and memory resources defined for a cluster.
type ResourceDescription struct {
	// +kubebuilder:validation:Pattern=`^(\d+m|\d+(\.\d{1,3})?)$`
	CPU string `json:"cpu"`

	// +kubebuilder:validation:Pattern=`^(\d+(e\d+)?|\d+(\.\d+)?(e\d+)?[EPTGMK]i?)$`
	Memory string `json:"memory"`

	// +kubebuilder:validation:Pattern=`^(\d+(e\d+)?|\d+(\.\d+)?(e\d+)?[EPTGMK]i?)$`
	EphemeralStorage string `json:"ephemeral-storage"`
}

// Resources describes requests and limits for the cluster resources.
type Resources struct {
	ResourceRequests ResourceDescription `json:"requests,omitempty"`
	ResourceLimits   ResourceDescription `json:"limits,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Environment{}, &EnvironmentList{})
}

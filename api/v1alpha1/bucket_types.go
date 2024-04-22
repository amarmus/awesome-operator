package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type BucketOnDeletePolicy string
type BucketCloud string

const (
	BucketOnDeletePolicyDestroy BucketOnDeletePolicy = "destroy"
	BucketOnDeletePolicyIgnore  BucketOnDeletePolicy = "ignore"

	BucketCloudGcp BucketCloud = "gcp"
)

// BucketSpec defines the desired state of Bucket
type BucketSpec struct {
	// Cloud platform
	// +kubebuilder:validation:Enum=gcp
	// +kubebuilder:validation:Required
	Cloud BucketCloud `json:"cloud"`

	// FullName is the cloud storage bucket full name
	// +kubebuilder:validation:Required
	FullName string `json:"fullName"`

	// OnDeletePolicy defines the behavior when the Deployment/Bucket objects are deleted
	// +kubebuilder:validation:Enum=destroy;ignore
	// +kubebuilder:validation:Required
	OnDeletePolicy BucketOnDeletePolicy `json:"onDeletePolicy"`
}

// BucketStatus defines the observed state of Bucket
type BucketStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// CreatedAt is the cloud storage bucket creation time
	CreatedAt string `json:"createdAt,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Bucket is the Schema for the buckets API
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketSpec   `json:"spec,omitempty"`
	Status BucketStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BucketList contains a list of Bucket
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bucket{}, &BucketList{})
}

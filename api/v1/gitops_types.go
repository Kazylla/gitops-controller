package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GitOpsSpec defines the desired state of GitOps
type GitOpsSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	AWSProfile string `json:"aws_profile,omitempty"`

	ImagePath      string `json:"image_path"`
	ImageTagFormat string `json:"image_tag_format"`

	GitRepo          string   `json:"git_repo"`
	GitBranch        string   `json:"git_branch"`
	GitReleaseBranch string   `json:"git_release_branch"`
	GitPaths         []string `json:"git_paths"`
	GitCommitName    string   `json:"git_commit_name"`
	GitCommitEmail   string   `json:"git_commit_email"`
}

// GitOpsStatus defines the observed state of GitOps
type GitOpsStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +optional
	CurrentTag string `json:"current_tag"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GitOps is the Schema for the gitops API
type GitOps struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitOpsSpec   `json:"spec,omitempty"`
	Status GitOpsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GitOpsList contains a list of GitOps
type GitOpsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitOps `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitOps{}, &GitOpsList{})
}

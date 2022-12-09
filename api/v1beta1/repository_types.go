package v1beta1

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	OperatorName    = "gitbackup"
	DefaultGitImage = "alpine/git:2.36.2"
)

// GetOwnedConfigMapName returns "gitbackup-repository-{r.Name}-gitconfig"
func (r Repository) GetOwnedConfigMapName() string {
	return strings.Join([]string{OperatorName, "repository", r.Name, "gitconfig"}, "-")
}

// GetOwnedCronJobName returns "gitbackup-{r.Name}"
func (r Repository) GetOwnedCronJobName() string {
	return strings.Join([]string{OperatorName, r.Name}, "-")
}

// RepositorySpec defines the desired state of Repository
type RepositorySpec struct {
	// Src specifies the source repository in URL format.
	Src string `json:"src"`
	// Dst specifies the destination repository in URL format.
	Dst string `json:"dst"`

	// Schedule in Cron format.
	Schedule string `json:"schedule"`
	// TimeZone in TZ database name.
	// See also: https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#time-zones
	// +optional
	TimeZone *string `json:"timeZone,omitempty"`

	// GitImage specifies the container image to run.
	// +optional
	GitImage *string `json:"gitImage,omitempty"`
	// ImagePullSecret specifies the name of the Secret in the same namespace used to pull the GitImage.
	// +optional
	ImagePullSecret *corev1.LocalObjectReference `json:"imagePullSecret,omitempty"`

	// GitConfig specifies the name of the configmap resource in the same namespace used to mount .git-config
	// Note that "[credential]\nhelper=store" is required to use GitCredentials.
	// +optional
	GitConfig *corev1.LocalObjectReference `json:"gitConfig,omitempty"`
	// GitCredentials specifies the name of the Secret in the same namespace used to mount .git-credentials
	// +optional
	GitCredentials *corev1.LocalObjectReference `json:"gitCredentials,omitempty"`
}

// RepositoryStatus defines the observed state of Repository
type RepositoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=repo;repos

// Repository is the Schema for the repositories API
type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositorySpec   `json:"spec,omitempty"`
	Status RepositoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RepositoryList contains a list of Repository
type RepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Repository{}, &RepositoryList{})
}

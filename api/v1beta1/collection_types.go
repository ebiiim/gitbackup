package v1beta1

import (
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CycleCronByMinuteInSameHour cycles cron minute.
// Assumes cronStr is "1 2 3 4 5" format.
// "30 6 * * *" -> "31 6 * * *" -> ... "59 6 * * *" -> "0 6 * * *" -> "1 6 * * *" -> ...
func CycleCronByMinuteInSameHour(cronStr string) (string, error) {
	ss := strings.Split(cronStr, " ")
	if len(ss) != 5 {
		return "", fmt.Errorf("cronStr must be \"1 2 3 4 5\" format but got %s", cronStr)
	}
	minute, err := strconv.Atoi(ss[0])
	if err != nil || minute < 0 || minute >= 60 {
		return "", fmt.Errorf("cronStr has invalid minute field cronStr=%s minute=%d err=%v", cronStr, minute, err)
	}
	minute = (minute + 1) % 60
	ss[0] = strconv.Itoa(minute)
	return strings.Join(ss, " "), nil
}

// GetOwnedConfigMapName returns "gitbackup-gitconfig-collection-{r.Name}"
func (r Collection) GetOwnedConfigMapName() string {
	return strings.Join([]string{OperatorName, "gitconfig", "collection", r.Name}, "-")
}

// GetOwnedRepositoryNames returns ["gitbackup-{r.Name}-{r.Repos[i].Name}", ...]
func (r Collection) GetOwnedRepositoryNames() []string {
	prefix := strings.Join([]string{OperatorName, r.Name, ""}, "-")
	names := make([]string, len(r.Spec.Repos))
	for i, cr := range r.Spec.Repos {
		var name string

		if cr.Name != nil {
			name = *cr.Name
		} else {
			// use the last element of cr.Src as name
			crSrc := strings.Split(cr.Src, "/")
			name = crSrc[len(crSrc)-1]
		}

		names[i] = prefix + name
	}
	return names
}

// CollectionSpec defines the desired state of Collection
type CollectionSpec struct {
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

	// Repos specifies repositories to backup.
	Repos []CollectionRepoURL `json:"repos"`
}

type CollectionRepoURL struct {
	// Name specifies the name for the repository. (default: the last element of `Src`)
	// +optional
	Name *string `json:"name,omitempty"`

	// Src specifies the source repository in URL format.
	Src string `json:"src"`
	// Dst specifies the destination repository in URL format.
	Dst string `json:"dst"`
}

// CollectionStatus defines the observed state of Collection
type CollectionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Collection is the Schema for the collections API
type Collection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CollectionSpec   `json:"spec,omitempty"`
	Status CollectionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CollectionList contains a list of Collection
type CollectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Collection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Collection{}, &CollectionList{})
}

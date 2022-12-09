package v1beta1

import (
	"errors"
	"strings"

	cron "github.com/robfig/cron/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var repositorylog = logf.Log.WithName("repository-resource")

func (r *Repository) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-gitbackup-ebiiim-com-v1beta1-repository,mutating=true,failurePolicy=fail,sideEffects=None,groups=gitbackup.ebiiim.com,resources=repositories,verbs=create;update,versions=v1beta1,name=mrepository.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Repository{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Repository) Default() {
	repositorylog.Info("default", "name", r.Name)

	if r.Spec.GitImage == nil {
		r.Spec.GitImage = pointer.String(DefaultGitImage)
	}
	if r.Spec.GitConfig == nil {
		r.Spec.GitConfig = &corev1.LocalObjectReference{Name: r.GetOwnedConfigMapName()}
	}
}

// NOTE: change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-gitbackup-ebiiim-com-v1beta1-repository,mutating=false,failurePolicy=fail,sideEffects=None,groups=gitbackup.ebiiim.com,resources=repositories,verbs=create;update,versions=v1beta1,name=vrepository.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Repository{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Repository) ValidateCreate() error {
	repositorylog.Info("validate create", "name", r.Name)

	if err := r.validateCron(); err != nil {
		return err
	}
	if err := r.validateURL(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Repository) ValidateUpdate(old runtime.Object) error {
	repositorylog.Info("validate update", "name", r.Name)

	if err := r.validateCron(); err != nil {
		return err
	}
	if err := r.validateURL(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
// NOTE: nothing to validate upon object deletion.
func (r *Repository) ValidateDelete() error { return nil }

func (r *Repository) validateCron() error {
	_, err := cron.ParseStandard(r.Spec.Schedule)
	return err
}

func (r *Repository) validateURL() error {
	if !isValidURLs(r.Spec.Src, r.Spec.Dst) {
		return errors.New("invalid src or dst URL")
	}
	return nil
}

func isValidURLs(s ...string) bool {
	for _, ss := range s {
		if strings.Contains(ss, " ") {
			return false
		}
	}
	return true
}

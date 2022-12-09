package v1beta1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var collectionlog = logf.Log.WithName("collection-resource")

func (r *Collection) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-gitbackup-ebiiim-com-v1beta1-collection,mutating=true,failurePolicy=fail,sideEffects=None,groups=gitbackup.ebiiim.com,resources=collections,verbs=create;update,versions=v1beta1,name=mcollection.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Collection{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Collection) Default() {
	collectionlog.Info("default", "name", r.Name)

	if r.Spec.GitConfig == nil {
		r.Spec.GitConfig = &corev1.LocalObjectReference{Name: r.GetOwnedConfigMapName()}
	}

}

// NOTE: change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-gitbackup-ebiiim-com-v1beta1-collection,mutating=false,failurePolicy=fail,sideEffects=None,groups=gitbackup.ebiiim.com,resources=collections,verbs=create;update,versions=v1beta1,name=vcollection.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Collection{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Collection) ValidateCreate() error {
	collectionlog.Info("validate create", "name", r.Name)

	if err := r.validateCron(); err != nil {
		return err
	}
	if err := r.validateRepos(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Collection) ValidateUpdate(old runtime.Object) error {
	collectionlog.Info("validate update", "name", r.Name)

	if err := r.validateCron(); err != nil {
		return err
	}
	if err := r.validateRepos(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
// NOTE: nothing to validate upon object deletion.
func (r *Collection) ValidateDelete() error { return nil }

func (r *Collection) validateCron() error {
	_, err := CycleCronByMinuteInSameHour(r.Spec.Schedule)
	return err
}

func (r *Collection) validateRepos() error {
	for i, cr := range r.Spec.Repos {
		if cr.Name != nil && len(validation.IsDNS1123Subdomain(*cr.Name)) != 0 {
			return fmt.Errorf("name must be RFC1123 DNS Subdomain string on spec.repos[%d]", i)
		}
		if !isValidURLSet(cr.Src, cr.Dst) {
			return fmt.Errorf("invalid src or dst URL on spec.repos[%d]", i)
		}
	}
	return nil
}

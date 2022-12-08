package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
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

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-gitbackup-ebiiim-com-v1beta1-collection,mutating=true,failurePolicy=fail,sideEffects=None,groups=gitbackup.ebiiim.com,resources=collections,verbs=create;update,versions=v1beta1,name=mcollection.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Collection{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Collection) Default() {
	collectionlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-gitbackup-ebiiim-com-v1beta1-collection,mutating=false,failurePolicy=fail,sideEffects=None,groups=gitbackup.ebiiim.com,resources=collections,verbs=create;update,versions=v1beta1,name=vcollection.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Collection{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Collection) ValidateCreate() error {
	collectionlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Collection) ValidateUpdate(old runtime.Object) error {
	collectionlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Collection) ValidateDelete() error {
	collectionlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

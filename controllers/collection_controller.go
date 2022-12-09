package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1beta1 "github.com/ebiiim/gitbackup/api/v1beta1"
)

// CollectionReconciler reconciles a Collection object
type CollectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gitbackup.ebiiim.com,resources=collections,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gitbackup.ebiiim.com,resources=collections/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gitbackup.ebiiim.com,resources=collections/finalizers,verbs=update

// Reconcile moves the current state of the cluster closer to the desired state.
func (r *CollectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	lg := log.FromContext(ctx)
	lg.Info("Reconcile")

	var col v1beta1.Collection
	err := r.Get(ctx, req.NamespacedName, &col)
	if errors.IsNotFound(err) {
		lg.Info("Collection is already deleted")
		return ctrl.Result{}, nil
	}
	if err != nil {
		lg.Error(err, "unable to get Collection")
		return ctrl.Result{}, err
	}
	if !col.DeletionTimestamp.IsZero() {
		lg.Info("Collection is being deleted")
		return ctrl.Result{}, nil
	}

	if err := r.reconcileGitConfig(ctx, col); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.reconcileGitCredentials(ctx, col); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.reconcileRepos(ctx, col); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CollectionReconciler) reconcileGitConfig(ctx context.Context, col v1beta1.Collection) error {
	lg := log.FromContext(ctx)
	lg.Info("reconcileGitConfig")

	cmName := col.GetOwnedConfigMapName()

	if col.Spec.GitConfig.Name != cmName {
		lg.Info("GitConfig specified")
		return nil
	}

	lg.Info("ensure default GitConfig cm created")

	cm := &corev1.ConfigMap{}
	cm.SetNamespace(col.Namespace)
	cm.SetName(cmName)

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, cm, func() error {
		cm.Data = map[string]string{
			".gitconfig": "[credential]\n\thelper = store",
		}
		return ctrl.SetControllerReference(&col, cm, r.Scheme)
	})

	if err != nil {
		// NOTE: A ConfigMap with the same name as the default GitConfig cm may exist
		lg.Error(err, "unable to create or update default GitConfig cm")
	}

	lg.Info("default GitConfig cm", "op", op)

	return nil
}

func (r *CollectionReconciler) reconcileGitCredentials(ctx context.Context, col v1beta1.Collection) error {
	lg := log.FromContext(ctx)
	lg.Info("reconcileGitCredentials")

	if col.Spec.GitCredentials != nil {
		lg.Info("GitCredentials specified")
		return nil
	}

	lg.Info("GitCredentials is nil")
	return nil
}

func sameOwner(a, b metav1.OwnerReference) bool {
	aa, errA := schema.ParseGroupVersion(a.APIVersion)
	bb, errB := schema.ParseGroupVersion(b.APIVersion)
	if errA != nil || errB != nil {
		return false
	}
	return (aa.Group == bb.Group) && (a.Kind == b.Kind) && (a.Name == b.Name)
}

func (r *CollectionReconciler) reconcileRepos(ctx context.Context, col v1beta1.Collection) error {
	lg := log.FromContext(ctx)
	lg.Info("reconcileRepos")

	desiredRepoNames := col.GetOwnedRepositoryNames()

	// put the list into a map to make search time O(1)
	desiredRepoNamesMap := make(map[string]struct{}, len(desiredRepoNames))
	for _, name := range desiredRepoNames {
		desiredRepoNamesMap[name] = struct{}{}
	}

	// ensure Repositories that are no longer needed are deleted
	var curRepos v1beta1.RepositoryList
	if err := r.List(ctx, &curRepos, &client.ListOptions{Namespace: col.Namespace}); err != nil {
		lg.Error(err, "unable to list Repositories")
	}
	for _, repo := range curRepos.Items {
		if !metav1.IsControlledBy(&repo, &col) {
			continue
		}
		if _, ok := desiredRepoNamesMap[repo.Name]; !ok {
			if err := r.Delete(ctx, &repo); err != nil {
				lg.Error(err, "unable to delete repo", "obj", repo)
			}
		}
	}

	// ensure Repositories created
	sched := col.Spec.Schedule
	for i, cr := range col.Spec.Repos {
		lg.Info("ensure Repository created", "name", desiredRepoNames[i])

		repo := &v1beta1.Repository{}
		repo.SetNamespace(col.Namespace)
		repo.SetName(desiredRepoNames[i])

		op, err := ctrl.CreateOrUpdate(ctx, r.Client, repo, func() error {
			repo.Spec = v1beta1.RepositorySpec{
				Src:             cr.Src,
				Dst:             cr.Dst,
				Schedule:        sched,
				TimeZone:        col.Spec.TimeZone,
				GitImage:        col.Spec.GitImage,
				ImagePullSecret: col.Spec.ImagePullSecret,
				GitConfig:       col.Spec.GitConfig,
				GitCredentials:  col.Spec.GitCredentials,
			}
			return ctrl.SetControllerReference(&col, repo, r.Scheme)
		})
		if err != nil {
			// NOTE: A Repository with the same name as desiredRepoNames[i] may exist
			lg.Error(err, "unable to create or update Repository", "name", desiredRepoNames[i])
		}
		// the cron expression is validated by Validating Webhook so no need to handle errors here
		sched, _ = v1beta1.CycleCronByMinuteInSameHour(sched)

		lg.Info("Repository reconciled", "name", desiredRepoNames[i], "op", op)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CollectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Collection{}).
		Owns(&v1beta1.Repository{}).
		Complete(r)
}

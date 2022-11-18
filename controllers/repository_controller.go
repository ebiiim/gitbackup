package controllers

import (
	"context"
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	batchv1apply "k8s.io/client-go/applyconfigurations/batch/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1beta1 "github.com/ebiiim/gitbackup/api/v1beta1"
)

const (
	ControllerName = v1beta1.OperatorName + "-repository-controller"
)

// RepositoryReconciler reconciles a Repository object
type RepositoryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gitbackup.ebiiim.com,resources=repositories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gitbackup.ebiiim.com,resources=repositories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gitbackup.ebiiim.com,resources=repositories/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;delete

// Reconcile moves the current state of the cluster closer to the desired state.
func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	lg := log.FromContext(ctx)
	lg.Info("Reconcile")

	var repo v1beta1.Repository
	err := r.Get(ctx, req.NamespacedName, &repo)
	if errors.IsNotFound(err) {
		lg.Info("Repository is already deleted")
		return ctrl.Result{}, nil
	}
	if err != nil {
		lg.Error(err, "unable to get Repository")
		return ctrl.Result{}, err
	}
	if !repo.DeletionTimestamp.IsZero() {
		lg.Info("Repository is being deleted")
		return ctrl.Result{}, nil
	}

	if err := r.reconcileGitConfig(ctx, repo); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.reconcileGitCredentials(ctx, repo); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.reconcileCronJob(ctx, repo); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *RepositoryReconciler) reconcileGitConfig(ctx context.Context, repo v1beta1.Repository) error {
	lg := log.FromContext(ctx)
	lg.Info("reconcileGitConfig")

	cmName := repo.GetOwnedConfigMapName()

	if repo.Spec.GitConfig.Name != cmName {
		lg.Info("GitConfig specified")
		return nil
	}

	lg.Info("ensure default GitConfig cm created")

	cm := &corev1.ConfigMap{}
	cm.SetNamespace(repo.Namespace)
	cm.SetName(cmName)

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, cm, func() error {
		cm.Data = map[string]string{
			".gitconfig": "[credential]\n\thelper = store",
		}
		return ctrl.SetControllerReference(&repo, cm, r.Scheme)
	})

	if err != nil {
		lg.Error(err, "unable to create or update default GitConfig cm")
	}

	lg.Info("default GitConfig cm", "op", op)

	return nil
}

func (r *RepositoryReconciler) reconcileGitCredentials(ctx context.Context, repo v1beta1.Repository) error {
	lg := log.FromContext(ctx)
	lg.Info("reconcileGitCredentials")

	if repo.Spec.GitCredentials != nil {
		lg.Info("GitCredentials specified")
		return nil
	}

	lg.Info("GitCredentials is nil")
	return nil
}

func (r *RepositoryReconciler) reconcileCronJob(ctx context.Context, repo v1beta1.Repository) error {
	lg := log.FromContext(ctx)
	lg.Info("reconcileCronJob")

	// generate shell commands
	srcs := strings.Split(repo.Spec.Src, "/")
	srcRepoName := srcs[len(srcs)-1]
	echo := func(format string, a ...any) string {
		logPrefix := "echo $(date -Iseconds) gitbackup: "
		return fmt.Sprintf(logPrefix+format, a...)
	}
	script := strings.Join([]string{
		echo("start"),
		echo("set .gitconfig"),
		"cp /gitconfig/.gitconfig /root/.gitconfig",
		echo("set .git-credentials"),
		"cp /gitcredentials/.git-credentials /root/.git-credentials",
		"set -e",
		echo("clone src repo '%s'", repo.Spec.Src),
		fmt.Sprintf("git clone --mirror '%s'", repo.Spec.Src),
		fmt.Sprintf("cd '%s.git'", srcRepoName),
		echo("push to dst repo '%s'", repo.Spec.Dst),
		fmt.Sprintf("git push --mirror '%s'", repo.Spec.Dst),
		"set +e",
		echo("completed"),
	}, ";")

	// create server-side apply config

	var volumes []*corev1apply.VolumeApplyConfiguration
	var volumeMounts []*corev1apply.VolumeMountApplyConfiguration

	volumes = append(volumes, corev1apply.Volume().
		WithName("gitconfig").
		WithConfigMap(corev1apply.ConfigMapVolumeSource().
			WithName(repo.Spec.GitConfig.Name).
			WithDefaultMode(256)),
	)
	volumeMounts = append(volumeMounts, corev1apply.VolumeMount().
		WithName("gitconfig").
		WithMountPath("/gitconfig"),
	)
	if repo.Spec.GitCredentials != nil {
		volumes = append(volumes, corev1apply.Volume().
			WithName("gitcredentials").
			WithSecret(corev1apply.SecretVolumeSource().
				WithSecretName(repo.Spec.GitCredentials.Name).
				WithDefaultMode(256)))
		volumeMounts = append(volumeMounts, corev1apply.VolumeMount().
			WithName("gitcredentials").
			WithMountPath("/gitcredentials"))
	}

	var containers []*corev1apply.ContainerApplyConfiguration

	containers = append(containers, corev1apply.Container().
		WithName("git").
		WithImage(*repo.Spec.GitImage).
		WithCommand(
			"/bin/sh",
			"-c",
			script,
		).
		WithVolumeMounts(volumeMounts...),
	)

	podTemplateSpec := corev1apply.PodTemplateSpec().WithSpec(corev1apply.PodSpec().
		WithRestartPolicy(corev1.RestartPolicyNever).
		WithContainers(containers...).
		WithVolumes(volumes...))
	if repo.Spec.ImagePullSecret != nil {
		podTemplateSpec.Spec.WithImagePullSecrets(corev1apply.LocalObjectReference().
			WithName(repo.Spec.ImagePullSecret.Name))
	}

	cronJobSpec := batchv1apply.CronJobSpec().
		WithSchedule(repo.Spec.Schedule).
		// Without this setting, CronJobs will stop working after 100 failures (including "suspend: true").
		WithStartingDeadlineSeconds(4 * 3600).
		// No need to backup concurrently and git commands can be cancelled.
		WithConcurrencyPolicy(batchv1.ReplaceConcurrent).
		WithJobTemplate(batchv1apply.JobTemplateSpec().WithSpec(batchv1apply.JobSpec().
			WithParallelism(1).
			WithCompletions(1).
			WithTemplate(podTemplateSpec)))
	if repo.Spec.TimeZone != nil {
		cronJobSpec.WithTimeZone(*repo.Spec.TimeZone)
	}

	gvk, err := apiutil.GVKForObject(&repo, r.Scheme)
	if err != nil {
		lg.Error(err, "unable to get GVK for Repository")
		return err
	}
	ownerReference := metav1apply.OwnerReference().
		WithAPIVersion(gvk.GroupVersion().Identifier()).
		WithKind(gvk.Kind).
		WithName(repo.Name).
		WithUID(repo.GetUID()).
		WithBlockOwnerDeletion(true).
		WithController(true)

	cronJob := batchv1apply.CronJob(repo.GetOwnedCronJobName(), repo.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       v1beta1.OperatorName,
			"app.kubernetes.io/instance":   repo.Name,
			"app.kubernetes.io/created-by": ControllerName,
		}).
		WithOwnerReferences(ownerReference).
		WithSpec(cronJobSpec)

	// do server-side apply
	// get current config > extract > not equal? > send patch

	var cur batchv1.CronJob
	if err := r.Get(ctx, client.ObjectKeyFromObject(&repo), &cur); err != nil && !errors.IsNotFound(err) {
		lg.Error(err, "unable to get current CronJob")
		return err
	}
	curApplyConfig, err := batchv1apply.ExtractCronJob(&cur, ControllerName)
	if err != nil {
		lg.Error(err, "unable to extract current CronJob")
		return err
	}
	if equality.Semantic.DeepEqual(cronJob, curApplyConfig) {
		lg.Info("no changes are made")
		return nil
	}
	lg.Info("do server-side apply")
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cronJob)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}
	if err := r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: ControllerName,
		Force:        pointer.Bool(true),
	}); err != nil {
		lg.Error(err, "unable to create or update CronJob")
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Repository{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}

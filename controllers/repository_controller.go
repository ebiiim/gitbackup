/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	"k8s.io/apimachinery/pkg/types"
	batchv1apply "k8s.io/client-go/applyconfigurations/batch/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1beta1 "github.com/ebiiim/gitbackup/api/v1beta1"
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

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Repository object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
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

func namespacedName(obj client.Object) types.NamespacedName {
	return types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}
}

func (r *RepositoryReconciler) reconcileGitConfig(ctx context.Context, repo v1beta1.Repository) error {
	lg := log.FromContext(ctx)
	lg.Info("reconcileGitConfig")

	cmName := v1beta1.DefaultGitConfigPrefix + repo.Name

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

	// create server-side apply config

	var volumes []*corev1apply.VolumeApplyConfiguration
	var volumeMounts []*corev1apply.VolumeMountApplyConfiguration

	volumes = append(volumes, corev1apply.Volume().
		WithName("gitconfig").
		WithConfigMap(corev1apply.ConfigMapVolumeSource().
			WithName(repo.Spec.GitConfig.Name)),
	)
	volumeMounts = append(volumeMounts, corev1apply.VolumeMount().
		WithName("gitconfig").
		WithMountPath("/root"),
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

	srcs := strings.Split(repo.Spec.Src, "/")
	srcRepoName := srcs[len(srcs)-1]
	containers = append(containers, corev1apply.Container().
		WithName("git").
		WithImage(*repo.Spec.GitImage).
		WithCommand(
			"/bin/sh",
			"-c",
			fmt.Sprintf("cp /gitcredentials/.git-credentials /root/.git-credentials ; clone --mirror %s ; cd %s.git ; git push --mirror %s", repo.Spec.Src, srcRepoName, repo.Spec.Dst),
		).
		WithVolumeMounts(volumeMounts...),
	)

	podTemplateSpec := corev1apply.PodTemplateSpec().WithSpec(corev1apply.PodSpec().
		WithRestartPolicy(corev1.RestartPolicyNever).
		WithContainers(containers...).
		WithVolumes(volumes...))

	cronJobSpec := batchv1apply.CronJobSpec().
		WithSchedule(repo.Spec.Schedule).
		WithJobTemplate(batchv1apply.JobTemplateSpec().WithSpec(batchv1apply.JobSpec().
			WithParallelism(1).
			WithCompletions(1).
			WithTemplate(podTemplateSpec)))
	if repo.Spec.TimeZone != nil {
		cronJobSpec.WithTimeZone(*repo.Spec.TimeZone)
	}

	cronJob := batchv1apply.CronJob(v1beta1.AppName+"-"+repo.Name, repo.Namespace).
		WithLabels(map[string]string{
			"app.kubernetes.io/name":       v1beta1.AppName,
			"app.kubernetes.io/instance":   repo.Name,
			"app.kubernetes.io/created-by": v1beta1.ControllerName,
		}).
		WithSpec(cronJobSpec)

	// do server-side apply
	// get current config > extract > not equal? > send patch

	var cur batchv1.CronJob
	if err := r.Get(ctx, namespacedName(&repo), &cur); err != nil && !errors.IsNotFound(err) {
		lg.Error(err, "unable to get current CronJob")
		return err
	}
	curApplyConfig, err := batchv1apply.ExtractCronJob(&cur, v1beta1.ControllerName)
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
		FieldManager: v1beta1.ControllerName,
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

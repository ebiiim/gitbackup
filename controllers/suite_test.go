package controllers_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	v1beta1 "github.com/ebiiim/gitbackup/api/v1beta1"
	"github.com/ebiiim/gitbackup/controllers"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = v1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var (
	waitShort = func() { time.Sleep(300 * time.Millisecond) }
	waitLong  = func() { time.Sleep(2000 * time.Millisecond) }
	testNS    = "default"
	testRepo1 = v1beta1.Repository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo1",
			Namespace: testNS,
		},
		Spec: v1beta1.RepositorySpec{
			Src:             "https://example.com/src",
			Dst:             "https://example.com/dst",
			Schedule:        "0 6 * * *",
			TimeZone:        nil,
			GitImage:        pointer.String(v1beta1.DefaultGitImage),
			ImagePullSecret: nil,
			GitConfig: &corev1.LocalObjectReference{
				Name: "gitbackup-gitconfig-test-repo1",
			},
			GitCredentials: nil,
		},
	}
	testRepo2 = v1beta1.Repository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo2",
			Namespace: testNS,
		},
		Spec: v1beta1.RepositorySpec{
			Src:      "https://example.com/src",
			Dst:      "https://example.com/dst",
			Schedule: "0 6 * * *",
			TimeZone: pointer.String("Asia/Tokyo"),
			GitImage: pointer.String(v1beta1.DefaultGitImage),
			ImagePullSecret: &corev1.LocalObjectReference{
				Name: "user-specified-image-pull-secret",
			},
			GitConfig: &corev1.LocalObjectReference{
				Name: "user-created-cm",
			},
			GitCredentials: &corev1.LocalObjectReference{
				Name: "user-specified-git-secret",
			},
		},
	}
)

var _ = Describe("Repository controller", func() {
	var cncl context.CancelFunc

	BeforeEach(func() {
		ctx, cancel := context.WithCancel(context.Background())
		cncl = cancel

		var err error
		err = k8sClient.DeleteAllOf(ctx, &v1beta1.Repository{}, client.InNamespace(testNS))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &batchv1.CronJob{}, client.InNamespace(testNS))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &corev1.ConfigMap{}, client.InNamespace(testNS))
		Expect(err).NotTo(HaveOccurred())
		waitShort()

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
		})
		Expect(err).NotTo(HaveOccurred())

		reconciler := controllers.RepositoryReconciler{
			Client: k8sClient,
			Scheme: scheme.Scheme,
		}
		err = reconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			err := mgr.Start(ctx)
			if err != nil {
				panic(err)
			}
		}()
		waitShort()
	})

	AfterEach(func() {
		cncl() // stop the mgr
		waitShort()
	})

	It("should create CronJob and ConfigMap", func() {
		repo := testRepo1
		ctx := context.Background()

		err := k8sClient.Create(ctx, &repo)
		Expect(err).NotTo(HaveOccurred())

		cm := corev1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: testNS, Name: repo.Spec.GitConfig.Name}, &cm)
		}).Should(Succeed())
		Expect(cm.Data).Should(HaveKey(".gitconfig"))

		cj := batchv1.CronJob{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: testNS, Name: v1beta1.OperatorName + "-" + repo.Name}, &cj)
		}).Should(Succeed())
		Expect(cj.Spec.Schedule).Should(Equal(repo.Spec.Schedule))
	})

	It("should only create CronJob", func() {
		repo := testRepo2
		ctx := context.Background()

		err := k8sClient.Create(ctx, &repo)
		Expect(err).NotTo(HaveOccurred())

		cm := corev1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: testNS, Name: repo.Spec.GitConfig.Name}, &cm)
		}).ShouldNot(Succeed())

		cj := batchv1.CronJob{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: testNS, Name: v1beta1.OperatorName + "-" + repo.Name}, &cj)
		}).Should(Succeed())
		Expect(cj.Spec.Schedule).Should(Equal(repo.Spec.Schedule))
	})
})

var (
	testColl1 = v1beta1.Collection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-coll1",
			Namespace: testNS,
		},
		Spec: v1beta1.CollectionSpec{
			Schedule:        "0 6 * * *",
			TimeZone:        nil,
			GitImage:        pointer.String(v1beta1.DefaultGitImage),
			ImagePullSecret: nil,
			GitConfig: &corev1.LocalObjectReference{
				Name: "gitbackup-gitconfig-collection-test-coll1",
			},
			GitCredentials: nil,
			Repos: []v1beta1.CollectionRepoURL{
				{
					Name: pointer.String("foo"),
					Src:  "http://example.com/src/foo",
					Dst:  "http://example.com/dst/foo",
				},
				{
					Name: pointer.String("bar"),
					Src:  "http://example.com/src/barbarbar",
					Dst:  "http://example.com/dst/barbarbar",
				},
				{
					Name: nil,
					Src:  "http://example.com/src/baz",
					Dst:  "http://example.com/dst/bazbazbaz",
				},
			},
		},
	}
	testColl2 = v1beta1.Collection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-coll2",
			Namespace: testNS,
		},
		Spec: v1beta1.CollectionSpec{
			Schedule: "0 6 * * *",
			TimeZone: pointer.String("Asia/Tokyo"),
			GitImage: pointer.String(v1beta1.DefaultGitImage),
			ImagePullSecret: &corev1.LocalObjectReference{
				Name: "user-specified-image-pull-secret",
			},
			GitConfig: &corev1.LocalObjectReference{
				Name: "user-created-cm",
			},
			GitCredentials: &corev1.LocalObjectReference{
				Name: "user-specified-git-secret",
			},
			Repos: []v1beta1.CollectionRepoURL{
				{
					Name: pointer.String("foo"),
					Src:  "http://example.com/src/foo",
					Dst:  "http://example.com/dst/foo",
				},
				{
					Name: pointer.String("bar"),
					Src:  "http://example.com/src/barbarbar",
					Dst:  "http://example.com/dst/barbarbar",
				},
				{
					Name: nil,
					Src:  "http://example.com/src/baz",
					Dst:  "http://example.com/dst/bazbazbaz",
				},
			},
		},
	}
	testColl3 = v1beta1.Collection{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-coll3",
			Namespace: testNS,
		},
		Spec: v1beta1.CollectionSpec{
			Schedule:        "0 6 * * *",
			TimeZone:        nil,
			GitImage:        pointer.String(v1beta1.DefaultGitImage),
			ImagePullSecret: nil,
			GitConfig: &corev1.LocalObjectReference{
				Name: "gitbackup-gitconfig-collection-test-coll3",
			},
			GitCredentials: nil,
			Repos:          []v1beta1.CollectionRepoURL{},
		},
	}
)

var _ = Describe("Collection controller", func() {
	var cncl context.CancelFunc

	BeforeEach(func() {
		ctx, cancel := context.WithCancel(context.Background())
		cncl = cancel

		var err error
		err = k8sClient.DeleteAllOf(ctx, &v1beta1.Collection{}, client.InNamespace(testNS))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &v1beta1.Repository{}, client.InNamespace(testNS))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &batchv1.CronJob{}, client.InNamespace(testNS))
		Expect(err).NotTo(HaveOccurred())
		err = k8sClient.DeleteAllOf(ctx, &corev1.ConfigMap{}, client.InNamespace(testNS))
		Expect(err).NotTo(HaveOccurred())
		waitShort()

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme: scheme.Scheme,
		})
		Expect(err).NotTo(HaveOccurred())

		reconciler := controllers.CollectionReconciler{
			Client: k8sClient,
			Scheme: scheme.Scheme,
		}
		err = reconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			err := mgr.Start(ctx)
			if err != nil {
				panic(err)
			}
		}()
		waitShort()
	})

	AfterEach(func() {
		cncl() // stop the mgr
		waitShort()
	})

	It("should create Repositories and a ConfigMap", func() {
		coll := testColl1
		ctx := context.Background()

		err := k8sClient.Create(ctx, &coll)
		Expect(err).NotTo(HaveOccurred())

		cm := corev1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: testNS, Name: coll.Spec.GitConfig.Name}, &cm)
		}).Should(Succeed())
		Expect(cm.Data).Should(HaveKey(".gitconfig"))

		for _, cr := range coll.GetOwnedRepositoryNames() {
			var repo v1beta1.Repository
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: testNS, Name: cr}, &repo)
			}).Should(Succeed())
			Expect(len(repo.OwnerReferences)).To(Equal(1))
		}
	})

	It("should only create Repositories", func() {
		coll := testColl2
		ctx := context.Background()

		err := k8sClient.Create(ctx, &coll)
		Expect(err).NotTo(HaveOccurred())

		cm := corev1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: testNS, Name: coll.Spec.GitConfig.Name}, &cm)
		}).ShouldNot(Succeed())

		for _, cr := range coll.GetOwnedRepositoryNames() {
			var repo v1beta1.Repository
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: testNS, Name: cr}, &repo)
			}).Should(Succeed())
			Expect(len(repo.OwnerReferences)).To(Equal(1))
		}
	})

	It("should only create a ConfigMap", func() {
		coll := testColl3
		ctx := context.Background()

		err := k8sClient.Create(ctx, &coll)
		Expect(err).NotTo(HaveOccurred())

		cm := corev1.ConfigMap{}
		Eventually(func() error {
			return k8sClient.Get(ctx, client.ObjectKey{Namespace: testNS, Name: coll.Spec.GitConfig.Name}, &cm)
		}).Should(Succeed())
		Expect(cm.Data).Should(HaveKey(".gitconfig"))

		for _, cr := range coll.GetOwnedRepositoryNames() {
			var repo v1beta1.Repository
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Namespace: testNS, Name: cr}, &repo)
			}).ShouldNot(Succeed())
		}
	})
})

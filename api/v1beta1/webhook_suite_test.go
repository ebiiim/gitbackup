package v1beta1_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	//+kubebuilder:scaffold:imports
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	v1beta1 "github.com/ebiiim/gitbackup/api/v1beta1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Webhook Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.Background())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "config", "webhook")},
		},
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	scheme := runtime.NewScheme()
	err = v1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = admissionv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	// start webhook server using Manager
	webhookInstallOptions := &testEnv.WebhookInstallOptions
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		Host:               webhookInstallOptions.LocalServingHost,
		Port:               webhookInstallOptions.LocalServingPort,
		CertDir:            webhookInstallOptions.LocalServingCertDir,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	Expect(err).NotTo(HaveOccurred())

	err = (&v1beta1.Repository{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&v1beta1.Collection{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:webhook

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
	}()

	// wait for the webhook server to get ready
	dialer := &net.Dialer{Timeout: time.Second}
	addrPort := fmt.Sprintf("%s:%d", webhookInstallOptions.LocalServingHost, webhookInstallOptions.LocalServingPort)
	Eventually(func() error {
		conn, err := tls.DialWithDialer(dialer, "tcp", addrPort, &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}).Should(Succeed())

})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Repository webhook", func() {
	dir := "testdata/repository"
	Context("mutating", func() {
		It("should mutate repositories", func() {
			testMutateRepository(mustOpen(dir, "mutate_minimal_before.yaml"), mustOpen(dir, "mutate_minimal_after.yaml"))
			testMutateRepository(mustOpen(dir, "mutate_all_before.yaml"), mustOpen(dir, "mutate_all_after.yaml"))
		})
	})
	Context("validating", func() {
		It("should create valid repositories", func() {
			want := true
			testValidateRepository(mustOpen(dir, "validate_all.yaml"), want)
			testValidateRepository(mustOpen(dir, "validate_minimal.yaml"), want)
			testValidateRepository(mustOpen(dir, "validate_cron_weekly.yaml"), want)
			testValidateRepository(mustOpen(dir, "validate_cron_sun.yaml"), want)
			_ = want
		})
		It("should not create invalid repositories", func() {
			want := false
			testValidateRepository(mustOpen(dir, "validate_wrong_cron.yaml"), want)
			testValidateRepository(mustOpen(dir, "validate_wrong_url_src.yaml"), want)
			testValidateRepository(mustOpen(dir, "validate_wrong_url_dst.yaml"), want)
			testValidateRepository(mustOpen(dir, "validate_wrong_url_eq.yaml"), want)
			_ = want
		})
	})
})

func testMutateRepository(rIn, rWant io.Reader) {
	ctx2 := context.Background()

	var in, got, want v1beta1.Repository

	err := yaml.NewYAMLOrJSONDecoder(rIn, 32).Decode(&in)
	Expect(err).NotTo(HaveOccurred())

	err = yaml.NewYAMLOrJSONDecoder(rWant, 32).Decode(&want)
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Create(ctx2, &in)
	Expect(err).NotTo(HaveOccurred())
	err = k8sClient.Get(ctx2, client.ObjectKeyFromObject(&in), &got)
	Expect(err).NotTo(HaveOccurred())

	Expect(got.Spec).Should(Equal(want.Spec))

	err = k8sClient.Delete(ctx2, &got)
	Expect(err).NotTo(HaveOccurred())
}

func testValidateRepository(rIn io.Reader, shouldBeValid bool) {
	ctx2 := context.Background()

	var in v1beta1.Repository

	err := yaml.NewYAMLOrJSONDecoder(rIn, 32).Decode(&in)
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Create(ctx2, &in)
	if shouldBeValid {
		Expect(err).NotTo(HaveOccurred(), "Data: %+v", &in)
	} else {
		Expect(err).To(HaveOccurred(), "Data: %#v", &in)
	}

	if shouldBeValid {
		err = k8sClient.Delete(ctx2, &in)
		Expect(err).NotTo(HaveOccurred())
	}
}

var _ = Describe("Collection webhook", func() {
	dir := "testdata/collection"
	Context("mutating", func() {
		It("should mutate collections", func() {
			testMutateCollection(mustOpen(dir, "mutate_minimal_before.yaml"), mustOpen(dir, "mutate_minimal_after.yaml"))
			testMutateCollection(mustOpen(dir, "mutate_all_before.yaml"), mustOpen(dir, "mutate_all_after.yaml"))
		})
	})
	Context("validating", func() {
		It("should create valid collections", func() {
			want := true
			testValidateCollection(mustOpen(dir, "validate_all.yaml"), want)
			testValidateCollection(mustOpen(dir, "validate_minimal.yaml"), want)
			testValidateCollection(mustOpen(dir, "validate_cron_sun.yaml"), want)
			_ = want
		})
		It("should not create invalid collections", func() {
			want := false
			testValidateCollection(mustOpen(dir, "validate_wrong_cron.yaml"), want)
			testValidateCollection(mustOpen(dir, "validate_wrong_cron_weekly.yaml"), want)
			testValidateCollection(mustOpen(dir, "validate_wrong_url_src.yaml"), want)
			testValidateCollection(mustOpen(dir, "validate_wrong_url_dst.yaml"), want)
			testValidateCollection(mustOpen(dir, "validate_wrong_url_eq.yaml"), want)
			testValidateCollection(mustOpen(dir, "validate_wrong_reponame.yaml"), want)
			_ = want
		})
	})
})

func testMutateCollection(rIn, rWant io.Reader) {
	ctx2 := context.Background()

	var in, got, want v1beta1.Collection

	err := yaml.NewYAMLOrJSONDecoder(rIn, 32).Decode(&in)
	Expect(err).NotTo(HaveOccurred())

	err = yaml.NewYAMLOrJSONDecoder(rWant, 32).Decode(&want)
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Create(ctx2, &in)
	Expect(err).NotTo(HaveOccurred())
	err = k8sClient.Get(ctx2, client.ObjectKeyFromObject(&in), &got)
	Expect(err).NotTo(HaveOccurred())

	Expect(got.Spec).Should(Equal(want.Spec))

	err = k8sClient.Delete(ctx2, &got)
	Expect(err).NotTo(HaveOccurred())
}

func testValidateCollection(rIn io.Reader, shouldBeValid bool) {
	ctx2 := context.Background()

	var in v1beta1.Collection

	err := yaml.NewYAMLOrJSONDecoder(rIn, 32).Decode(&in)
	Expect(err).NotTo(HaveOccurred())

	err = k8sClient.Create(ctx2, &in)
	if shouldBeValid {
		Expect(err).NotTo(HaveOccurred(), "Data: %+v", &in)
	} else {
		Expect(err).To(HaveOccurred(), "Data: %#v", &in)
	}

	if shouldBeValid {
		err = k8sClient.Delete(ctx2, &in)
		Expect(err).NotTo(HaveOccurred())
	}
}

func mustOpen(filePath ...string) io.Reader {
	f, err := os.Open(filepath.Join(filePath...))
	if err != nil {
		panic(err)
	}
	return f
}

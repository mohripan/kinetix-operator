package controller

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	pipelinev1alpha1 "github.com/kinetix/kinetix-operator/operator/api/v1alpha1"
)

func TestPipelineReconcileWithEnvtest(t *testing.T) {
	if os.Getenv("KINETIX_RUN_ENVTEST") != "1" {
		t.Skip("set KINETIX_RUN_ENVTEST=1 and KUBEBUILDER_ASSETS to run envtest")
	}

	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("add client-go scheme: %v", err)
	}
	if err := pipelinev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add pipeline scheme: %v", err)
	}

	env := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "config", "crd", "bases"),
			filepath.Join("testdata", "crds"),
		},
		ErrorIfCRDPathMissing: true,
	}
	cfg, err := env.Start()
	if err != nil {
		t.Fatalf("start envtest: %v", err)
	}
	t.Cleanup(func() {
		if err := env.Stop(); err != nil {
			t.Fatalf("stop envtest: %v", err)
		}
	})

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	pipeline := testPipeline("envtest")
	if err := k8sClient.Create(ctx, pipeline); err != nil {
		t.Fatalf("create pipeline: %v", err)
	}

	reconciler := &PipelineReconciler{
		Client:          k8sClient,
		Scheme:          scheme,
		BrokerBootstrap: "kafka:9092",
		StrimziCluster:  "kinetix-kafka",
	}
	if _, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: pipeline.Name, Namespace: pipeline.Namespace}}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	var deploy appsv1.Deployment
	if err := k8sClient.Get(ctx, NamespacedName("kinetix", "envtest-worker"), &deploy); err != nil {
		t.Fatalf("expected deployment: %v", err)
	}
	var cm corev1.ConfigMap
	if err := k8sClient.Get(ctx, NamespacedName("kinetix", "envtest-config"), &cm); err != nil {
		t.Fatalf("expected config map: %v", err)
	}
}

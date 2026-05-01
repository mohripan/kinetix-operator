package controller

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	pipelinev1alpha1 "github.com/kinetix/kinetix-operator/operator/api/v1alpha1"
)

func TestPipelineReconcileCreatesWorkerResources(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)
	pipeline := testPipeline("example")
	reconciler := testReconciler(t, scheme, pipeline)

	if _, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: pipeline.Name, Namespace: pipeline.Namespace}}); err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}

	var cm corev1.ConfigMap
	if err := reconciler.Get(ctx, NamespacedName("kinetix", "example-config"), &cm); err != nil {
		t.Fatalf("expected config map: %v", err)
	}
	if got := cm.Data["KINETIX_INPUT_TOPIC"]; got != "input" {
		t.Fatalf("KINETIX_INPUT_TOPIC = %q, want input", got)
	}
	if got := cm.Data["KINETIX_OUTPUT_TOPIC"]; got != "output" {
		t.Fatalf("KINETIX_OUTPUT_TOPIC = %q, want output", got)
	}
	if got := cm.Data["KINETIX_BROKERS"]; got != "kafka:9092" {
		t.Fatalf("KINETIX_BROKERS = %q, want kafka:9092", got)
	}
	if !metav1.IsControlledBy(&cm, pipeline) {
		t.Fatalf("config map is not controlled by pipeline")
	}

	var deploy appsv1.Deployment
	if err := reconciler.Get(ctx, NamespacedName("kinetix", "example-worker"), &deploy); err != nil {
		t.Fatalf("expected worker deployment: %v", err)
	}
	if deploy.Spec.Replicas == nil || *deploy.Spec.Replicas != 2 {
		t.Fatalf("replicas = %v, want 2", deploy.Spec.Replicas)
	}
	if got := deploy.Spec.Template.Spec.Containers[0].Image; got != "kinetix/normalize:v0.1.0" {
		t.Fatalf("image = %q, want kinetix/normalize:v0.1.0", got)
	}
	if got := deploy.Spec.Template.Spec.Containers[0].EnvFrom[0].ConfigMapRef.Name; got != "example-config" {
		t.Fatalf("config map ref = %q, want example-config", got)
	}
	if got := deploy.Spec.Template.ObjectMeta.Annotations["dapr.io/enabled"]; got != "true" {
		t.Fatalf("dapr annotation = %q, want true", got)
	}
	if deploy.Spec.Template.Spec.SecurityContext == nil || deploy.Spec.Template.Spec.SecurityContext.RunAsNonRoot == nil || !*deploy.Spec.Template.Spec.SecurityContext.RunAsNonRoot {
		t.Fatalf("worker pod must run as non-root")
	}
	if deploy.Spec.Template.Spec.SecurityContext.RunAsUser == nil || *deploy.Spec.Template.Spec.SecurityContext.RunAsUser != 10001 {
		t.Fatalf("worker pod runAsUser = %v, want 10001", deploy.Spec.Template.Spec.SecurityContext.RunAsUser)
	}
	if deploy.Spec.Template.Spec.SecurityContext.SeccompProfile == nil || deploy.Spec.Template.Spec.SecurityContext.SeccompProfile.Type != corev1.SeccompProfileTypeRuntimeDefault {
		t.Fatalf("worker pod seccomp profile = %#v, want RuntimeDefault", deploy.Spec.Template.Spec.SecurityContext.SeccompProfile)
	}
	containerSecurity := deploy.Spec.Template.Spec.Containers[0].SecurityContext
	if containerSecurity == nil || containerSecurity.AllowPrivilegeEscalation == nil || *containerSecurity.AllowPrivilegeEscalation {
		t.Fatalf("worker container must disallow privilege escalation")
	}
	if containerSecurity.ReadOnlyRootFilesystem == nil || !*containerSecurity.ReadOnlyRootFilesystem {
		t.Fatalf("worker container must use a read-only root filesystem")
	}
	if containerSecurity.Capabilities == nil || len(containerSecurity.Capabilities.Drop) != 1 || containerSecurity.Capabilities.Drop[0] != "ALL" {
		t.Fatalf("worker container capabilities drop = %#v, want ALL", containerSecurity.Capabilities)
	}
	if !metav1.IsControlledBy(&deploy, pipeline) {
		t.Fatalf("deployment is not controlled by pipeline")
	}

	var sourceTopic unstructured.Unstructured
	sourceTopic.SetAPIVersion("kafka.strimzi.io/v1beta2")
	sourceTopic.SetKind("KafkaTopic")
	if err := reconciler.Get(ctx, NamespacedName("kinetix", "example-source"), &sourceTopic); err != nil {
		t.Fatalf("expected source KafkaTopic: %v", err)
	}
	topicName, _, err := unstructured.NestedString(sourceTopic.Object, "spec", "topicName")
	if err != nil || topicName != "input" {
		t.Fatalf("source KafkaTopic spec.topicName = %q, err %v; want input", topicName, err)
	}

	var updated pipelinev1alpha1.Pipeline
	if err := reconciler.Get(ctx, NamespacedName("kinetix", "example"), &updated); err != nil {
		t.Fatalf("expected updated pipeline: %v", err)
	}
	condition := findCondition(updated.Status.Conditions, pipelinev1alpha1.ConditionReady)
	if condition == nil || condition.Status != metav1.ConditionTrue {
		t.Fatalf("Ready condition = %#v, want true", condition)
	}
}

func TestPipelineReconcileUpdatesExistingDeployment(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)
	pipeline := testPipeline("example")
	reconciler := testReconciler(t, scheme, pipeline)
	key := types.NamespacedName{Name: pipeline.Name, Namespace: pipeline.Namespace}

	if _, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: key}); err != nil {
		t.Fatalf("first reconcile failed: %v", err)
	}

	var stored pipelinev1alpha1.Pipeline
	if err := reconciler.Get(ctx, key, &stored); err != nil {
		t.Fatalf("expected pipeline: %v", err)
	}
	stored.Spec.Transforms[0].Image = "kinetix/normalize:v0.2.0"
	replicas := int32(3)
	stored.Spec.Transforms[0].Replicas = &replicas
	if err := reconciler.Update(ctx, &stored); err != nil {
		t.Fatalf("update pipeline: %v", err)
	}

	if _, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: key}); err != nil {
		t.Fatalf("second reconcile failed: %v", err)
	}

	var deploy appsv1.Deployment
	if err := reconciler.Get(ctx, NamespacedName("kinetix", "example-worker"), &deploy); err != nil {
		t.Fatalf("expected worker deployment: %v", err)
	}
	if got := deploy.Spec.Template.Spec.Containers[0].Image; got != "kinetix/normalize:v0.2.0" {
		t.Fatalf("image = %q, want kinetix/normalize:v0.2.0", got)
	}
	if deploy.Spec.Replicas == nil || *deploy.Spec.Replicas != 3 {
		t.Fatalf("replicas = %v, want 3", deploy.Spec.Replicas)
	}
}

func TestPipelineReconcileReportsInvalidSpec(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)
	pipeline := testPipeline("invalid")
	pipeline.Spec.Transforms = nil
	reconciler := testReconciler(t, scheme, pipeline)

	if _, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: pipeline.Name, Namespace: pipeline.Namespace}}); err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}

	var updated pipelinev1alpha1.Pipeline
	if err := reconciler.Get(ctx, NamespacedName("kinetix", "invalid"), &updated); err != nil {
		t.Fatalf("expected updated pipeline: %v", err)
	}
	condition := findCondition(updated.Status.Conditions, pipelinev1alpha1.ConditionReady)
	if condition == nil {
		t.Fatalf("expected Ready condition")
	}
	if condition.Status != metav1.ConditionFalse || condition.Reason != pipelinev1alpha1.ReasonInvalidSpec {
		t.Fatalf("Ready condition = %#v, want InvalidSpec false", condition)
	}
}

func TestPipelineReconcileDeleteCleansOwnedResourcesAndFinalizer(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)
	deletionTime := metav1.Now()
	pipeline := testPipeline("example")
	pipeline.Finalizers = []string{PipelineFinalizer}
	pipeline.DeletionTimestamp = &deletionTime
	pipeline.ResourceVersion = "1"

	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "example-config", Namespace: "kinetix"}}
	deploy := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "example-worker", Namespace: "kinetix"}}
	sourceTopic := kafkaTopicObject("kinetix", "example-source")
	sinkTopic := kafkaTopicObject("kinetix", "example-sink")
	reconciler := testReconciler(t, scheme, pipeline, cm, deploy, sourceTopic, sinkTopic)

	if _, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: pipeline.Name, Namespace: pipeline.Namespace}}); err != nil {
		t.Fatalf("reconcile delete failed: %v", err)
	}

	var deletedCM corev1.ConfigMap
	if err := reconciler.Get(ctx, NamespacedName("kinetix", "example-config"), &deletedCM); err == nil {
		t.Fatalf("expected config map to be deleted")
	}
	var deletedDeploy appsv1.Deployment
	if err := reconciler.Get(ctx, NamespacedName("kinetix", "example-worker"), &deletedDeploy); err == nil {
		t.Fatalf("expected deployment to be deleted")
	}

	var updated pipelinev1alpha1.Pipeline
	if err := reconciler.Get(ctx, NamespacedName("kinetix", "example"), &updated); err != nil {
		if apierrors.IsNotFound(err) {
			return
		}
		t.Fatalf("get pipeline after finalizer removal: %v", err)
	}
	for _, finalizer := range updated.Finalizers {
		if finalizer == PipelineFinalizer {
			t.Fatalf("expected finalizer to be removed")
		}
	}
}

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("add client-go scheme: %v", err)
	}
	if err := pipelinev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add pipeline scheme: %v", err)
	}
	return scheme
}

func testReconciler(t *testing.T, scheme *runtime.Scheme, objects ...client.Object) *PipelineReconciler {
	t.Helper()
	builder := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objects...).
		WithStatusSubresource(&pipelinev1alpha1.Pipeline{})

	return &PipelineReconciler{
		Client:          builder.Build(),
		Scheme:          scheme,
		BrokerBootstrap: "kafka:9092",
		StrimziCluster:  "kinetix-kafka",
	}
}

func testPipeline(name string) *pipelinev1alpha1.Pipeline {
	replicas := int32(2)
	return &pipelinev1alpha1.Pipeline{
		TypeMeta: metav1.TypeMeta{
			APIVersion: pipelinev1alpha1.GroupVersion.String(),
			Kind:       "Pipeline",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  "kinetix",
			Generation: 1,
		},
		Spec: pipelinev1alpha1.PipelineSpec{
			Source: pipelinev1alpha1.EndpointSpec{
				Kind:   "KafkaSource",
				Topic:  "input",
				Schema: "example-v1",
			},
			Transforms: []pipelinev1alpha1.TransformSpec{{
				Name:     "normalize",
				Image:    "kinetix/normalize:v0.1.0",
				Replicas: &replicas,
			}},
			Sink: pipelinev1alpha1.EndpointSpec{
				Kind:  "KafkaSink",
				Topic: "output",
			},
		},
	}
}

func findCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}

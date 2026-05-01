package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	pipelinev1alpha1 "github.com/kinetix/kinetix-operator/operator/api/v1alpha1"
)

const (
	PipelineFinalizer = "pipeline.kinetix.io/finalizer"

	defaultBrokerBootstrap = "kinetix-kafka-kafka-bootstrap.kinetix.svc.cluster.local:9092"
	defaultWorkerImage     = "kinetix/demo:dev"
	defaultStrimziCluster  = "kinetix-kafka"
)

type PipelineReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	BrokerBootstrap string
	WorkerImage     string
	StrimziCluster  string
}

func (r *PipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var pipeline pipelinev1alpha1.Pipeline
	if err := r.Get(ctx, req.NamespacedName, &pipeline); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !pipeline.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, r.reconcileDelete(ctx, &pipeline)
	}

	if controllerutil.AddFinalizer(&pipeline, PipelineFinalizer) {
		if err := r.Update(ctx, &pipeline); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := validatePipeline(&pipeline); err != nil {
		setReadyCondition(&pipeline, metav1.ConditionFalse, pipelinev1alpha1.ReasonInvalidSpec, err.Error())
		if statusErr := r.updateStatus(ctx, &pipeline); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{}, nil
	}

	if err := r.reconcileResources(ctx, &pipeline); err != nil {
		logger.Error(err, "failed to reconcile pipeline resources")
		setReadyCondition(&pipeline, metav1.ConditionFalse, pipelinev1alpha1.ReasonReconcileFailed, err.Error())
		_ = r.updateStatus(ctx, &pipeline)
		return ctrl.Result{}, err
	}

	pipeline.Status.ObservedGeneration = pipeline.Generation
	pipeline.Status.WorkerDeployment = workerDeploymentName(&pipeline)
	pipeline.Status.ConfigMap = configMapName(&pipeline)
	pipeline.Status.SourceTopic = kafkaTopicResourceName(&pipeline, "source")
	pipeline.Status.SinkTopic = kafkaTopicResourceName(&pipeline, "sink")
	setReadyCondition(&pipeline, metav1.ConditionTrue, pipelinev1alpha1.ReasonResourcesReady, "Pipeline resources are reconciled")

	return ctrl.Result{}, r.updateStatus(ctx, &pipeline)
}

func (r *PipelineReconciler) reconcileResources(ctx context.Context, pipeline *pipelinev1alpha1.Pipeline) error {
	if err := r.applyConfigMap(ctx, pipeline); err != nil {
		return err
	}
	if err := r.applyWorkerDeployment(ctx, pipeline); err != nil {
		return err
	}
	if err := r.applyKafkaTopic(ctx, pipeline, "source", pipeline.Spec.Source.Topic); err != nil {
		return err
	}
	return r.applyKafkaTopic(ctx, pipeline, "sink", pipeline.Spec.Sink.Topic)
}

func (r *PipelineReconciler) reconcileDelete(ctx context.Context, pipeline *pipelinev1alpha1.Pipeline) error {
	if !controllerutil.ContainsFinalizer(pipeline, PipelineFinalizer) {
		return nil
	}

	for _, obj := range []client.Object{
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: workerDeploymentName(pipeline), Namespace: pipeline.Namespace}},
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: configMapName(pipeline), Namespace: pipeline.Namespace}},
		kafkaTopicObject(pipeline.Namespace, kafkaTopicResourceName(pipeline, "source")),
		kafkaTopicObject(pipeline.Namespace, kafkaTopicResourceName(pipeline, "sink")),
	} {
		if err := r.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}

	controllerutil.RemoveFinalizer(pipeline, PipelineFinalizer)
	return r.Update(ctx, pipeline)
}

func (r *PipelineReconciler) applyConfigMap(ctx context.Context, pipeline *pipelinev1alpha1.Pipeline) error {
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: configMapName(pipeline), Namespace: pipeline.Namespace}}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, cm, func() error {
		cm.Labels = labelsForPipeline(pipeline)
		cm.Data = r.configData(pipeline)
		return controllerutil.SetControllerReference(pipeline, cm, r.Scheme)
	})
	return err
}

func (r *PipelineReconciler) applyWorkerDeployment(ctx context.Context, pipeline *pipelinev1alpha1.Pipeline) error {
	deploy := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: workerDeploymentName(pipeline), Namespace: pipeline.Namespace}}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deploy, func() error {
		transform := pipeline.Spec.Transforms[0]
		replicas := int32(1)
		if transform.Replicas != nil && *transform.Replicas > 0 {
			replicas = *transform.Replicas
		}

		labels := labelsForPipeline(pipeline)
		deploy.Labels = labels
		deploy.Spec.Replicas = &replicas
		deploy.Spec.Selector = &metav1.LabelSelector{MatchLabels: selectorLabelsForPipeline(pipeline)}
		deploy.Spec.Template.ObjectMeta.Labels = labels
		deploy.Spec.Template.ObjectMeta.Annotations = map[string]string{
			"dapr.io/enabled":      "true",
			"dapr.io/app-id":       pipeline.Name + "-worker",
			"prometheus.io/scrape": "true",
			"prometheus.io/port":   "8080",
			"prometheus.io/path":   "/metrics",
		}
		deploy.Spec.Template.Spec.Containers = []corev1.Container{{
			Name:            "worker",
			Image:           imageForTransform(transform, r.workerImage()),
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"/worker"},
			EnvFrom: []corev1.EnvFromSource{{
				ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: configMapName(pipeline)},
				},
			}},
			Ports: []corev1.ContainerPort{{
				Name:          "metrics",
				ContainerPort: 8080,
			}},
			Resources: resourcesForTransform(transform),
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{Path: "/metrics", Port: intstr.FromString("metrics")},
				},
				InitialDelaySeconds: 5,
				PeriodSeconds:       10,
			},
		}}
		return controllerutil.SetControllerReference(pipeline, deploy, r.Scheme)
	})
	return err
}

func (r *PipelineReconciler) applyKafkaTopic(ctx context.Context, pipeline *pipelinev1alpha1.Pipeline, role, topicName string) error {
	topic := kafkaTopicObject(pipeline.Namespace, kafkaTopicResourceName(pipeline, role))
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, topic, func() error {
		topic.SetLabels(map[string]string{
			"app.kubernetes.io/name":       "kinetix-topic",
			"app.kubernetes.io/part-of":    "kinetix",
			"app.kubernetes.io/managed-by": "kinetix-operator",
			"kinetix.io/pipeline":          pipeline.Name,
			"kinetix.io/topic-role":        role,
			"strimzi.io/cluster":           r.strimziCluster(),
		})
		topic.Object["spec"] = map[string]any{
			"topicName":  topicName,
			"partitions": int64(1),
			"replicas":   int64(1),
			"config": map[string]any{
				"retention.ms": "604800000",
			},
		}
		return controllerutil.SetControllerReference(pipeline, topic, r.Scheme)
	})
	return err
}

func (r *PipelineReconciler) configData(pipeline *pipelinev1alpha1.Pipeline) map[string]string {
	transform := pipeline.Spec.Transforms[0]
	return map[string]string{
		"KINETIX_BROKERS":         r.brokerBootstrap(),
		"KINETIX_INPUT_TOPIC":     pipeline.Spec.Source.Topic,
		"KINETIX_OUTPUT_TOPIC":    pipeline.Spec.Sink.Topic,
		"KINETIX_GROUP":           pipeline.Name + "-worker",
		"KINETIX_SOURCE_KIND":     pipeline.Spec.Source.Kind,
		"KINETIX_SOURCE_SCHEMA":   pipeline.Spec.Source.Schema,
		"KINETIX_SINK_KIND":       pipeline.Spec.Sink.Kind,
		"KINETIX_TRANSFORM_NAME":  transform.Name,
		"KINETIX_TRANSFORM_IMAGE": imageForTransform(transform, r.workerImage()),
	}
}

func (r *PipelineReconciler) updateStatus(ctx context.Context, pipeline *pipelinev1alpha1.Pipeline) error {
	pipeline.Status.ObservedGeneration = pipeline.Generation
	if err := r.Status().Update(ctx, pipeline); err != nil {
		return err
	}
	return nil
}

func (r *PipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pipelinev1alpha1.Pipeline{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

func validatePipeline(pipeline *pipelinev1alpha1.Pipeline) error {
	var problems []string
	if pipeline.Spec.Source.Kind != "KafkaSource" {
		problems = append(problems, "spec.source.kind must be KafkaSource")
	}
	if pipeline.Spec.Source.Topic == "" {
		problems = append(problems, "spec.source.topic is required")
	}
	if pipeline.Spec.Sink.Kind != "KafkaSink" {
		problems = append(problems, "spec.sink.kind must be KafkaSink")
	}
	if pipeline.Spec.Sink.Topic == "" {
		problems = append(problems, "spec.sink.topic is required")
	}
	if len(pipeline.Spec.Transforms) == 0 {
		problems = append(problems, "spec.transforms must include at least one transform")
	}
	for i, transform := range pipeline.Spec.Transforms {
		if transform.Name == "" {
			problems = append(problems, fmt.Sprintf("spec.transforms[%d].name is required", i))
		}
		if errs := validation.IsDNS1123Label(transform.Name); transform.Name != "" && len(errs) > 0 {
			problems = append(problems, fmt.Sprintf("spec.transforms[%d].name must be a DNS-1123 label", i))
		}
		if transform.Image == "" {
			problems = append(problems, fmt.Sprintf("spec.transforms[%d].image is required", i))
		}
		if transform.Replicas != nil && *transform.Replicas < 1 {
			problems = append(problems, fmt.Sprintf("spec.transforms[%d].replicas must be at least 1", i))
		}
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func setReadyCondition(pipeline *pipelinev1alpha1.Pipeline, status metav1.ConditionStatus, reason, message string) {
	apimeta.SetStatusCondition(&pipeline.Status.Conditions, metav1.Condition{
		Type:               pipelinev1alpha1.ConditionReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: pipeline.Generation,
	})
}

func labelsForPipeline(pipeline *pipelinev1alpha1.Pipeline) map[string]string {
	labels := selectorLabelsForPipeline(pipeline)
	labels["app.kubernetes.io/part-of"] = "kinetix"
	labels["app.kubernetes.io/managed-by"] = "kinetix-operator"
	return labels
}

func selectorLabelsForPipeline(pipeline *pipelinev1alpha1.Pipeline) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name": "kinetix-worker",
		"kinetix.io/pipeline":    pipeline.Name,
	}
}

func configMapName(pipeline *pipelinev1alpha1.Pipeline) string {
	return pipeline.Name + "-config"
}

func workerDeploymentName(pipeline *pipelinev1alpha1.Pipeline) string {
	return pipeline.Name + "-worker"
}

func kafkaTopicResourceName(pipeline *pipelinev1alpha1.Pipeline, role string) string {
	return pipeline.Name + "-" + role
}

func kafkaTopicObject(namespace, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "kafka.strimzi.io/v1beta2",
			"kind":       "KafkaTopic",
			"metadata": map[string]any{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
}

func imageForTransform(transform pipelinev1alpha1.TransformSpec, fallback string) string {
	if transform.Image != "" {
		return transform.Image
	}
	return fallback
}

func resourcesForTransform(transform pipelinev1alpha1.TransformSpec) corev1.ResourceRequirements {
	if transform.Resources.Requests != nil || transform.Resources.Limits != nil {
		return transform.Resources
	}
	return corev1.ResourceRequirements{}
}

func (r *PipelineReconciler) brokerBootstrap() string {
	if r.BrokerBootstrap != "" {
		return r.BrokerBootstrap
	}
	return defaultBrokerBootstrap
}

func (r *PipelineReconciler) workerImage() string {
	if r.WorkerImage != "" {
		return r.WorkerImage
	}
	return defaultWorkerImage
}

func (r *PipelineReconciler) strimziCluster() string {
	if r.StrimziCluster != "" {
		return r.StrimziCluster
	}
	return defaultStrimziCluster
}

func NamespacedName(namespace, name string) types.NamespacedName {
	return types.NamespacedName{Namespace: namespace, Name: name}
}

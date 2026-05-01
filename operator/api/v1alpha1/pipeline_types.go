package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ConditionReady = "Ready"

	ReasonReconciling     = "Reconciling"
	ReasonInvalidSpec     = "InvalidSpec"
	ReasonResourcesReady  = "ResourcesReady"
	ReasonReconcileFailed = "ReconcileFailed"
)

type PipelineSpec struct {
	Source     EndpointSpec    `json:"source"`
	Transforms []TransformSpec `json:"transforms"`
	Sink       EndpointSpec    `json:"sink"`
}

type EndpointSpec struct {
	Kind   string `json:"kind"`
	Topic  string `json:"topic"`
	Schema string `json:"schema,omitempty"`
}

type TransformSpec struct {
	Name      string                      `json:"name"`
	Image     string                      `json:"image"`
	Replicas  *int32                      `json:"replicas,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

type PipelineStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	WorkerDeployment   string             `json:"workerDeployment,omitempty"`
	ConfigMap          string             `json:"configMap,omitempty"`
	SourceTopic        string             `json:"sourceTopic,omitempty"`
	SinkTopic          string             `json:"sinkTopic,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

func (in *Pipeline) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(Pipeline)
	in.DeepCopyInto(out)
	return out
}

func (in *Pipeline) DeepCopyInto(out *Pipeline) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = *in.Spec.DeepCopy()
	out.Status = *in.Status.DeepCopy()
}

func (in *PipelineSpec) DeepCopy() *PipelineSpec {
	if in == nil {
		return nil
	}
	out := new(PipelineSpec)
	*out = *in
	if in.Transforms != nil {
		out.Transforms = make([]TransformSpec, len(in.Transforms))
		for i := range in.Transforms {
			out.Transforms[i] = *in.Transforms[i].DeepCopy()
		}
	}
	return out
}

func (in *TransformSpec) DeepCopy() *TransformSpec {
	if in == nil {
		return nil
	}
	out := new(TransformSpec)
	*out = *in
	if in.Replicas != nil {
		replicas := *in.Replicas
		out.Replicas = &replicas
	}
	out.Resources = *in.Resources.DeepCopy()
	return out
}

func (in *PipelineStatus) DeepCopy() *PipelineStatus {
	if in == nil {
		return nil
	}
	out := new(PipelineStatus)
	*out = *in
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		copy(out.Conditions, in.Conditions)
	}
	return out
}

func (in *PipelineList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(PipelineList)
	in.DeepCopyInto(out)
	return out
}

func (in *PipelineList) DeepCopyInto(out *PipelineList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]Pipeline, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

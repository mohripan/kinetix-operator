package v1alpha1

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type PipelineValidator struct{}

func (PipelineValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	pipeline, ok := obj.(*Pipeline)
	if !ok {
		return nil, fmt.Errorf("expected Pipeline, got %T", obj)
	}
	return nil, ValidatePipelineSpec(&pipeline.Spec)
}

func (PipelineValidator) ValidateUpdate(_ context.Context, _ runtime.Object, newObj runtime.Object) (admission.Warnings, error) {
	pipeline, ok := newObj.(*Pipeline)
	if !ok {
		return nil, fmt.Errorf("expected Pipeline, got %T", newObj)
	}
	return nil, ValidatePipelineSpec(&pipeline.Spec)
}

func (PipelineValidator) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func ValidatePipelineSpec(spec *PipelineSpec) error {
	var problems []string
	if spec.Source.Kind != "KafkaSource" {
		problems = append(problems, "spec.source.kind must be KafkaSource")
	}
	if spec.Source.Topic == "" {
		problems = append(problems, "spec.source.topic is required")
	}
	if spec.Sink.Kind != "KafkaSink" {
		problems = append(problems, "spec.sink.kind must be KafkaSink")
	}
	if spec.Sink.Topic == "" {
		problems = append(problems, "spec.sink.topic is required")
	}
	if len(spec.Transforms) == 0 {
		problems = append(problems, "spec.transforms must include at least one transform")
	}
	for i, transform := range spec.Transforms {
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

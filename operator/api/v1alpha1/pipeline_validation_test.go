package v1alpha1

import (
	"strings"
	"testing"
)

func TestValidatePipelineSpecAcceptsPhase2Shape(t *testing.T) {
	replicas := int32(1)
	spec := &PipelineSpec{
		Source: EndpointSpec{Kind: "KafkaSource", Topic: "input", Schema: "user-event-v1"},
		Transforms: []TransformSpec{{
			Name:     "normalize",
			Image:    "kinetix/normalize:v0.1.0",
			Replicas: &replicas,
		}},
		Sink: EndpointSpec{Kind: "KafkaSink", Topic: "output"},
	}

	if err := ValidatePipelineSpec(spec); err != nil {
		t.Fatalf("ValidatePipelineSpec() error = %v, want nil", err)
	}
}

func TestValidatePipelineSpecRejectsInvalidSecurityRelevantFields(t *testing.T) {
	replicas := int32(0)
	spec := &PipelineSpec{
		Source: EndpointSpec{Kind: "HTTPSource"},
		Transforms: []TransformSpec{{
			Name:     "Bad_Name",
			Replicas: &replicas,
		}},
		Sink: EndpointSpec{Kind: "KafkaSink"},
	}

	err := ValidatePipelineSpec(spec)
	if err == nil {
		t.Fatalf("ValidatePipelineSpec() error = nil, want validation error")
	}
	for _, want := range []string{
		"spec.source.kind must be KafkaSource",
		"spec.source.topic is required",
		"spec.transforms[0].name must be a DNS-1123 label",
		"spec.transforms[0].image is required",
		"spec.transforms[0].replicas must be at least 1",
		"spec.sink.topic is required",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("ValidatePipelineSpec() = %q, want substring %q", err.Error(), want)
		}
	}
}

package v1alpha1

import "k8s.io/utils/pointer"

func (s *ServerlessSpec) Default() {
	if s.DockerRegistry == nil {
		s.DockerRegistry = &DockerRegistry{}
	}

	if s.DockerRegistry.EnableInternal == nil {
		s.DockerRegistry.EnableInternal = pointer.Bool(true)
	}
}

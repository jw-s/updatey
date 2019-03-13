package k8s

import (
	"github.com/jw-s/updatey/pkg/client/docker"
	"github.com/jw-s/updatey/pkg/version"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

// Interface defines the functionality related to kubernetes.
type Interface interface {
	GetPatches(ar *v1beta1.AdmissionRequest) (patches []*JSONPatch, err error)
	GetImagePullSecrets(imagePullSecrets []corev1.LocalObjectReference, namespace string) ([]*corev1.Secret, error)
}

// Wrapper is a simple helper type to wrap the kubernetes client.
type Wrapper struct {
	secretRetriever SecretInterface
	resolver        version.Resolver
	dockerClient    docker.Interface
}

// New returns a new Wrapper.
func New(secretRetriever SecretInterface, resolver version.Resolver, dockerClient docker.Interface) *Wrapper {
	return &Wrapper{
		secretRetriever: secretRetriever,
		resolver:        resolver,
		dockerClient:    dockerClient,
	}
}

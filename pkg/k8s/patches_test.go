package k8s

import (
	"encoding/json"
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jw-s/updatey/pkg/client/docker"
	"github.com/jw-s/updatey/pkg/version"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type testDockerClient struct {
	tags [][]string
	errs []error
}

func (c *testDockerClient) Tags(auth *docker.Auth, repository string) ([]string, error) {
	var tags []string
	var err error
	if len(c.tags) == 1 && len(c.errs) == 1 {
		return c.tags[0], c.errs[0]
	} else if len(c.tags) > 1 && len(c.errs) > 1 {
		tags, c.tags = c.tags[0], c.tags[1:]
		err, c.errs = c.errs[0], c.errs[1:]
		return tags, err
	} else {
		panic("testing error, no more elements to expect")
	}
}

type testResolver struct {
	resolve string
}

func (r *testResolver) Resolve(constraint string, tags []string) string {
	return r.resolve
}

func TestGetPatches(t *testing.T) {
	tests := []struct {
		o               runtime.Object
		dockerClient    docker.Interface
		resolver        version.Resolver
		secretRetriever SecretInterface
		expected        []*JSONPatch
		err             error
	}{
		{
			o: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind: "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "alpine:latest",
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind: "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "alpine:latest",
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind: "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "a",
						},
						{
							Name: "b",
						},
					},
					Containers: []corev1.Container{
						{
							Image: "alpine:latest",
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{
				secrets: map[string]*corev1.Secret{
					"a": &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "a",
						},
					},
					"b": newDockerSecret(corev1.DockerConfigKey, "docker.io", base64EncodedCredentials),
				},
				errs: map[string]error{
					"a": nil,
					"b": nil,
				},
				namespace: "default",
			},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind: "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Image: "alpine:latest",
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind: "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Image: "alpine:latest",
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind: "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Image: "alpine:latest",
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &corev1.ReplicationController{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicationController",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.ReplicationControllerSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &corev1.ReplicationController{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicationController",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.ReplicationControllerSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &corev1.ReplicationController{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicationController",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.ReplicationControllerSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &corev1.ReplicationController{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicationController",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.ReplicationControllerSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &corev1.ReplicationController{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicationController",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.ReplicationControllerSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &corev1.ReplicationController{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicationController",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.ReplicationControllerSpec{
					Template: &corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &batchv1.Job{
				TypeMeta: metav1.TypeMeta{
					Kind: "Job",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &batchv1.Job{
				TypeMeta: metav1.TypeMeta{
					Kind: "Job",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &batchv1.Job{
				TypeMeta: metav1.TypeMeta{
					Kind: "Job",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &batchv1.Job{
				TypeMeta: metav1.TypeMeta{
					Kind: "Job",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &batchv1.Job{
				TypeMeta: metav1.TypeMeta{
					Kind: "Job",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &batchv1.Job{
				TypeMeta: metav1.TypeMeta{
					Kind: "Job",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.ReplicaSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicaSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.ReplicaSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.ReplicaSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicaSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.ReplicaSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.ReplicaSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicaSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.ReplicaSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.ReplicaSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicaSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.ReplicaSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.ReplicaSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicaSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.ReplicaSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.ReplicaSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicaSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.ReplicaSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.StatefulSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.StatefulSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &appsv1.DaemonSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							InitContainers: []corev1.Container{
								{
									Image: "alpine:latest",
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &batchv1beta1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind: "CronJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: batchv1beta1.CronJobSpec{
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Image: "alpine:latest",
										},
									},
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/jobTemplate/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &batchv1beta1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind: "CronJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: batchv1beta1.CronJobSpec{
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Image: "alpine:latest",
										},
									},
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/jobTemplate/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &batchv1beta1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind: "CronJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: batchv1beta1.CronJobSpec{
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Image: "alpine:latest",
										},
									},
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/jobTemplate/spec/template/spec/containers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &batchv1beta1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind: "CronJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: batchv1beta1.CronJobSpec{
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									InitContainers: []corev1.Container{
										{
											Image: "alpine:latest",
										},
									},
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{
						"1.0",
					},
				},
				errs: []error{
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/jobTemplate/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &batchv1beta1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind: "CronJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: batchv1beta1.CronJobSpec{
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									InitContainers: []corev1.Container{
										{
											Image: "alpine:latest",
										},
									},
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/jobTemplate/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &batchv1beta1.CronJob{
				TypeMeta: metav1.TypeMeta{
					Kind: "CronJob",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: batchv1beta1.CronJobSpec{
					JobTemplate: batchv1beta1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									InitContainers: []corev1.Container{
										{
											Image: "alpine:latest",
										},
									},
								},
							},
						},
					},
				},
			},
			dockerClient: &testDockerClient{
				tags: [][]string{
					{},
					{
						"1.0",
					},
				},
				errs: []error{
					errors.New("not-valid-secret"),
					nil,
				},
			},
			resolver: &testResolver{
				resolve: "1.0",
			},
			secretRetriever: &testSecretRetriever{},
			expected: []*JSONPatch{
				&JSONPatch{
					Op:    "replace",
					Path:  "/spec/jobTemplate/spec/template/spec/initContainers/0/image",
					Value: "alpine:1.0",
				},
			},
		},
		{
			o: &corev1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind: "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		w := New(test.secretRetriever, test.resolver, test.dockerClient)

		patches, err := w.GetPatches(createAdmissionRequest(test.o))

		assert.Equal(t, test.err, err)

		assert.ElementsMatch(t, test.expected, patches)
	}
}

func createAdmissionRequest(o runtime.Object) *v1beta1.AdmissionRequest {

	b, err := json.Marshal(o)

	if err != nil {
		panic(err)
	}

	return &v1beta1.AdmissionRequest{
		Kind: metav1.GroupVersionKind{
			Kind: o.GetObjectKind().GroupVersionKind().Kind,
		},
		Object: runtime.RawExtension{
			Raw: b,
		},
	}
}

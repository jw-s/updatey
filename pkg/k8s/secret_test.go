package k8s

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testUser                 = "testuser"
	testPass                 = "testpass"
	base64EncodedCredentials = "dGVzdHVzZXI6dGVzdHBhc3M=" // testuser:testpass
)

type testSecretRetriever struct {
	secrets   map[string]*corev1.Secret
	namespace string
	errs      map[string]error
}

func (s *testSecretRetriever) Get(namespace, name string) (*corev1.Secret, error) {
	secret, exists := s.secrets[name]
	if !exists || namespace != s.namespace {
		return nil, &apierrors.StatusError{
			ErrStatus: metav1.Status{
				Reason: metav1.StatusReasonNotFound,
			},
		}
	}
	return secret, s.errs[name]
}

func TestExtractFromDockerSecret(t *testing.T) {
	tests := []struct {
		secretType   string
		domain       string
		image        string
		auth         string
		expectedUser string
		expectedPass string
		expectErr    bool
	}{
		{
			secretType: "notvalidtype",
			domain:     "docker.io",
			image:      "alpine",
			auth:       base64EncodedCredentials,
			expectErr:  true,
		},
		{
			secretType:   corev1.DockerConfigJsonKey,
			domain:       "docker.io",
			image:        "alpine:latest",
			auth:         base64EncodedCredentials,
			expectedUser: testUser,
			expectedPass: testPass,
		},
		{
			secretType: corev1.DockerConfigJsonKey,
			domain:     "docker.io",
			image:      "quay.io/alpine:latest",
			auth:       base64EncodedCredentials,
			expectErr:  true, // No valid image secret for quay.io
		},
		{
			secretType:   corev1.DockerConfigKey,
			domain:       "docker.io",
			image:        "alpine:latest",
			auth:         base64EncodedCredentials,
			expectedUser: testUser,
			expectedPass: testPass,
		},
		{
			secretType: corev1.DockerConfigKey,
			domain:     "docker.io",
			image:      "quay.io/alpine:latest",
			auth:       base64EncodedCredentials,
			expectErr:  true, // No valid image secret for quay.io
		},
	}

	for _, test := range tests {
		secret := newDockerSecret(test.secretType, test.domain, test.auth)
		username, password, err := ExtractFromDockerSecret(secret, test.image)

		if err != nil && !test.expectErr {
			t.Fatal(err)
		}

		assert.Equal(t, test.expectedUser, username)
		assert.Equal(t, test.expectedPass, password)

	}
}

func TestGetImagePullSecrets(t *testing.T) {
	tests := []struct {
		secretRetriever  *testSecretRetriever
		imagePullSecrets []corev1.LocalObjectReference
		namespace        string
		expected         []*corev1.Secret
		err              error
	}{
		{
			secretRetriever: &testSecretRetriever{
				secrets: map[string]*corev1.Secret{
					"mysecret": &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "mysecret",
						},
					},
				},
				namespace: "default",
			},
			imagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: "mysecret",
				},
			},
			namespace: "",
			expected: []*corev1.Secret{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mysecret",
					},
				},
			},
		},
		{
			secretRetriever: &testSecretRetriever{
				secrets: map[string]*corev1.Secret{
					"mysecret": &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "mysecret",
						},
					},
				},
				namespace: "default",
			},
			imagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: "mysecret",
				},
			},
			namespace: "mynamespace",
			expected:  []*corev1.Secret{},
		},
		{
			secretRetriever: &testSecretRetriever{
				secrets: map[string]*corev1.Secret{
					"mysecret": &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "mysecret",
						},
					},
				},
				namespace: "default",
			},
			imagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: "mysecret",
				},
			},
			namespace: "mynamespace",
			expected:  []*corev1.Secret{},
		},
		{
			secretRetriever: &testSecretRetriever{
				secrets: map[string]*corev1.Secret{
					"mysecret": &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "mysecret",
						},
					},
					"a": &corev1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name: "a",
						},
					},
				},
				errs: map[string]error{
					"a": errors.New("test error"),
				},
				namespace: "default",
			},
			imagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: "a",
				},
				{
					Name: "mysecret",
				},
			},

			namespace: "default",
			expected: []*corev1.Secret{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name: "mysecret",
					},
				},
			},
		},
	}

	for _, test := range tests {
		w := New(test.secretRetriever, nil, nil)

		secrets, err := w.GetImagePullSecrets(test.imagePullSecrets, test.namespace)

		assert.Equal(t, test.err, err)
		assert.ElementsMatch(t, test.expected, secrets)
	}
}

func newDockerSecret(secretType, domain, auth string) *corev1.Secret {
	var template string
	if secretType == corev1.DockerConfigJsonKey {
		template = `{
			"auths": {
			  "%s": {
				"auth": "%s",
				"email": ""
			  }
			}
		  }`
	} else {
		template = `{
			"%s": {
			  "auth": "%s",
			  "email": ""
			}
		  }`
	}

	return &corev1.Secret{
		Data: map[string][]byte{
			secretType: []byte(fmt.Sprintf(template, domain, auth)),
		},
	}
}

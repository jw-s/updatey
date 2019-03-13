package k8s

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/glog"

	"github.com/jw-s/updatey/pkg/client/docker"

	"github.com/docker/distribution/reference"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

// SecretInterface defines how to fetch secrets.
type SecretInterface interface {
	Get(namespace, name string) (*corev1.Secret, error)
}

// SecretRetriever provides functionality to fetch secrets.
type SecretRetriever struct {
	secretsGetter v1.SecretsGetter
}

// Get returns a namespaced secret of the specified name and possibily an error.
func (s *SecretRetriever) Get(namespace, name string) (*corev1.Secret, error) {
	return s.secretsGetter.Secrets(namespace).Get(name, metav1.GetOptions{})
}

// NewSecretRetriever creates a new SecretRetriever
func NewSecretRetriever(secretsGetter v1.SecretsGetter) *SecretRetriever {
	return &SecretRetriever{
		secretsGetter: secretsGetter,
	}
}

type registryConfigs struct {
	Auths registryConfig `json:"auths,omitempty"`
}

type registryConfig map[string]struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Auth     string `json:"auth,omitempty"`
}

// GetImagePullSecrets returns a slice of secrets for and possibily an error.
func (w *Wrapper) GetImagePullSecrets(imagePullSecrets []corev1.LocalObjectReference, namespace string) (secrets []*corev1.Secret, err error) {
	if namespace == "" {
		namespace = "default"
	}

	for _, pullSecret := range imagePullSecrets {
		secret, err := w.secretRetriever.Get(namespace, pullSecret.Name)

		if err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			glog.Error(err)
			continue
		}
		secrets = append(secrets, secret)
	}
	return secrets, nil

}

// ExtractFromDockerSecret returns a username and password for a docker registry from the secret and possibly an error.
func ExtractFromDockerSecret(secret *corev1.Secret, image string) (username, password string, err error) {

	var registryConfigs registryConfigs
	config, exists := secret.Data[corev1.DockerConfigJsonKey]

	if !exists {
		config, exists = secret.Data[corev1.DockerConfigKey]

		if !exists {
			err = errors.New("no docker config found in secret")
			return username, password, err
		}
		var registryConfig registryConfig
		if err = json.Unmarshal(config, &registryConfig); err != nil {
			return username, password, err
		}

		registryConfigs.Auths = registryConfig
	} else {
		if err = json.Unmarshal(config, &registryConfigs); err != nil {
			return username, password, err
		}
	}
	base, _, err := docker.Split(image)

	if err != nil {
		return username, password, err
	}

	imageName, err := reference.ParseNormalizedNamed(base)

	if err != nil {
		return username, password, err
	}

	registryDomain := reference.Domain(imageName)

	registryConfig, exists := registryConfigs.Auths[registryDomain]

	if !exists {
		return username, password, fmt.Errorf("registry domain: %s does not exist", registryDomain)
	}

	b, err := base64.StdEncoding.DecodeString(registryConfig.Auth)

	if err != nil {
		return username, password, err
	}

	basicAuth := strings.Split(string(b), ":")

	if len(basicAuth) != 2 {
		return username, password, errors.New("auth field should equal basic auth syntax")
	}

	return basicAuth[0], basicAuth[1], nil

}

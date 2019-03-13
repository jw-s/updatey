package k8s

import (
	"encoding/json"
	"fmt"

	"github.com/golang/glog"
	"github.com/jw-s/updatey/pkg/client/docker"

	"k8s.io/api/admission/v1beta1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	podSpecPath      = "/spec"
	templateSpecPath = "/spec/template/spec"
	cronJobSpecPath  = "/spec/jobTemplate/spec/template/spec"
)

// JSONPatch is the type which stores the json patch format as per http://jsonpatch.com.
type JSONPatch struct {
	From  string      `json:"from,omitempty"`
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

type templateKind struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              struct {
		Template struct {
			Spec corev1.PodSpec `json:"spec,omitempty"`
		} `json:"template,omitempty"`
	} `json:"spec,omitempty"`
}

// GetPatches returns a slice of json patches based on the admission request and possibily an error.
func (w *Wrapper) GetPatches(ar *v1beta1.AdmissionRequest) (patches []*JSONPatch, err error) {
	var (
		spec      *corev1.PodSpec
		specPath  string
		namespace string
	)

	switch ar.Kind.Kind {
	case "Pod":
		pod := corev1.Pod{}
		err = json.Unmarshal(ar.Object.Raw, &pod)
		if err != nil {
			return nil, err
		}
		spec, specPath, namespace = &pod.Spec, podSpecPath, pod.Namespace
	case "ReplicationController", "Job", "ReplicaSet", "Deployment", "StatefulSet", "DaemonSet":
		templateKind := templateKind{}
		err = json.Unmarshal(ar.Object.Raw, &templateKind)
		if err != nil {
			return nil, err
		}
		spec, specPath, namespace = &templateKind.Spec.Template.Spec, templateSpecPath, templateKind.Namespace
	case "CronJob":
		cronJob := batchv1beta1.CronJob{}
		err = json.Unmarshal(ar.Object.Raw, &cronJob)
		if err != nil {
			return nil, err
		}
		spec, specPath, namespace = &cronJob.Spec.JobTemplate.Spec.Template.Spec, cronJobSpecPath, cronJob.Namespace
	default:
		return nil, nil
	}

	return w.processPodSpec(spec, specPath, namespace)
}

func (w *Wrapper) processPodSpec(podSpec *corev1.PodSpec, specPath, namespace string) (patches []*JSONPatch, err error) {
	for _, containerType := range []string{"initContainers", "containers"} {
		var containers []corev1.Container
		switch containerType {
		case "initContainers":
			containers = podSpec.InitContainers
		case "containers":
			containers = podSpec.Containers
		}

		secrets, err := w.GetImagePullSecrets(podSpec.ImagePullSecrets, namespace)

		if err != nil {
			return patches, err
		}

	containerLoop:
		for containerIndex, container := range containers {

			repository, tag, err := docker.Split(container.Image)
			if err != nil {
				glog.Error(err)
				continue containerLoop
			}

			tags, err := w.dockerClient.Tags(nil, repository)

			if err != nil {
			secretLoop:
				for _, secret := range secrets {
					username, password, err := ExtractFromDockerSecret(secret, container.Image)
					if err != nil {
						glog.Error(err)
						continue secretLoop
					}

					tags, err = w.dockerClient.Tags(&docker.Auth{
						Username: username,
						Password: password,
					}, repository)

					if err != nil {
						glog.Error(err)
						continue secretLoop
					}

					break secretLoop
				}
			}
			newImageVersion := w.resolver.Resolve(tag, tags)

			patches = append(patches, &JSONPatch{
				Op:    "replace",
				Path:  fmt.Sprintf("%s/%s/%v/image", specPath, containerType, containerIndex),
				Value: fmt.Sprintf("%s:%s", repository, newImageVersion),
			})
		}
	}
	return patches, nil
}

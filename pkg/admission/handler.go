package admission

import (
	"encoding/json"
	"net/http"

	"github.com/jw-s/updatey/pkg/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/api/admission/v1beta1"
)

// AdmitHandler is the mutating webhook handler.
func AdmitHandler(client k8s.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		if req.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var ar v1beta1.AdmissionReview
		if err := json.NewDecoder(req.Body).Decode(&ar); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("unable to decode body"))
			return
		}

		patches, err := client.GetPatches(ar.Request)

		if err != nil {
			ar.Response = &v1beta1.AdmissionResponse{
				UID:     ar.Request.UID,
				Allowed: false,
				Result: &v1.Status{
					Message: err.Error(),
				},
			}

			resp, err := json.Marshal(ar)

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("unable to encode response"))
				return
			}

			w.Write(resp)
			return
		}

		jsonPatch, err := json.Marshal(patches)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("unable to encode patches"))
			return
		}
		pt := v1beta1.PatchTypeJSONPatch
		ar.Response = &v1beta1.AdmissionResponse{
			UID:       ar.Request.UID,
			Allowed:   true,
			PatchType: &pt,
			Patch:     jsonPatch,
		}

		resp, err := json.Marshal(ar)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("unable to encode response"))
			return
		}

		w.Write(resp)

	}
}

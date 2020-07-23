package secretfixer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
	v1beta1 "k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var logger *zap.Logger

func writeError(w http.ResponseWriter, errStr string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf(`{ "error": "%s" }`, errStr)))
}

func writeAdmissionReviewError(w http.ResponseWriter, ar *v1beta1.AdmissionReview, errStr string) {
	ar.Response = &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: errStr,
		},
	}
	e := json.NewEncoder(w)
	err := e.Encode(ar)
	if err != nil {
		logger.Error("unable to encode error response", zap.Error(err))
		http.Error(w, "unable to encode error response", http.StatusInternalServerError)
	}
}

type patch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

func mutate(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logger.Error("wrong content-type", zap.String("contentType", contentType))
		writeError(w, fmt.Sprintf("wrong content type: %s", contentType))
		return
	}

	var admissionReview v1beta1.AdmissionReview
	b := &strings.Builder{}
	tee := io.TeeReader(r.Body, b)
	d := json.NewDecoder(tee)
	defer r.Body.Close()
	err := d.Decode(&admissionReview)
	if err != nil {
		logger.Error("unable to decode body", zap.Error(err), zap.String("body", b.String()))
		writeError(w, err.Error())
		return
	}

	logger.Info("decoded body", zap.Error(err), zap.String("body", b.String()))

	if admissionReview.Request.Kind.Kind != "Secret" {
		logger.Error("wrong kind", zap.String("kind", admissionReview.Request.Kind.Kind))
		writeAdmissionReviewError(w, &admissionReview, fmt.Sprintf("wrong kind: %s", admissionReview.Request.Kind.Kind))
		return
	}

	var secret v1.Secret
	err = json.Unmarshal(admissionReview.Request.Object.Raw, &secret)
	if err != nil {
		logger.Error("unable to decode object", zap.Error(err), zap.ByteString("objectRaw", admissionReview.Request.Object.Raw))
		writeAdmissionReviewError(w, &admissionReview, "unable to decode object")
		return
	}

	var patchBuf []byte
	if secret.Type == v1.SecretTypeTLS {
		if _, exists := secret.Data["ca.crt"]; !exists {
			var patches []patch

			p := patch{}
			p.Op = "add"
			p.Path = "data/ca.crt"
			p.Value = "DEADBEEFBADCAFE"

			patches = append(patches, p)

			patchBuf, err = json.Marshal(patches)
			if err != nil {
				logger.Error("unable to marshal patch", zap.Error(err), zap.ByteString("patch", patchBuf))
				writeAdmissionReviewError(w, &admissionReview, "unable to marshal patch")
				return
			}
		}
	}

	admissionReview.Response = &v1beta1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
	}
	if len(patchBuf) > 0 {
		admissionReview.Response.Patch = patchBuf
	}

	buf, err := json.Marshal(admissionReview)
	if err != nil {
		logger.Error("unable to encode response", zap.Error(err))
		writeError(w, fmt.Sprintf("unable to encode response: %s", err))
	}
	logger.Info("admission response", zap.ByteString("Response", buf))
	w.Write(buf)
}

func Handler(l *zap.Logger) http.Handler {
	logger = l

	r := chi.NewMux()
	r.Post("/", mutate)

	return r
}

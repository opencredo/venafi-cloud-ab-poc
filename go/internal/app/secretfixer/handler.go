package ledger

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var logger *zap.Logger

func writeError(w http.ResponseWriter, errStr string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf(`{ "error": "%s" }`, errStr)))
}

func writeAdmissionReviewError(w http.ResponseWriter, ar *v1.AdmissionReview, errStr string) {
	ar.Response = &v1.AdmissionResponse{
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

func mutate(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logger.Error("wrong content-type", zap.String("contentType", contentType))
		writeError(w, fmt.Sprintf("wrong content type: %s", contentType))
		return
	}

	var admissionReview *v1.AdmissionReview
	d := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := d.Decode(admissionReview)
	if err != nil {
		logger.Error("unable to decode body", zap.Error(err))
		writeError(w, err.Error())
		return
	}

	if admissionReview.Request.Kind.Kind != "Secret" {
		logger.Error("wrong kind", zap.String("kind", admissionReview.Request.Kind.Kind))
		writeAdmissionReviewError(w, admissionReview, fmt.Sprintf("wrong kind: %s", admissionReview.Request.Kind.Kind))
		return
	}

	admissionReview.Response = &v1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
	}

	e := json.NewEncoder(w)
	err = e.Encode(admissionReview)
	if err != nil {
		logger.Error("unable to encode response", zap.Error(err))
		writeError(w, fmt.Sprintf("unable to encode response: %s", err))
	}
}

func Handler(l *zap.Logger) http.Handler {
	logger = l

	r := chi.NewMux()
	r.Get("/", mutate)

	return r
}

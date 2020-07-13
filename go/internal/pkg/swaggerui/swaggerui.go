package swaggerui

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/ledgerserver"
	"github.com/rakyll/statik/fs"
)

func Handler() http.Handler {
	s, _ := ledgerserver.GetSwagger()

	r := chi.NewMux()
	r.MethodFunc(http.MethodGet, "/swaggerui/swagger.json", func(w http.ResponseWriter, req *http.Request) {
		serviceName := req.Header.Get("X-Ocvab-Service")
		namespace := req.Header.Get("X-Ocvab-Namespace")
		url := "/"
		if serviceName != "" && namespace != "" {
			url = "/" + namespace + "/" + serviceName + "/"
		}
		s.Servers = []*openapi3.Server{
			{
				URL:         url,
				Description: "This service",
				Variables:   s.Servers[0].Variables,
			},
		}
		s.AddServer(&openapi3.Server{})
		b, _ := s.MarshalJSON()
		w.Write(b)
	})

	fs, _ := fs.New()
	fileServer(r, fs)

	return r
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, root http.FileSystem) {
	r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
		isFileName, _ := regexp.MatchString(`^.*/[^./]*\..*[^/]$`, req.RequestURI)
		if isFileName || strings.HasSuffix(req.RequestURI, "/") {
			rctx := chi.RouteContext(req.Context())
			pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
			fs := http.StripPrefix(pathPrefix, http.FileServer(root))
			fs.ServeHTTP(w, req)
		} else {
			redirectDest := req.Header.Get("X-Original-Uri")
			if redirectDest == "" {
				redirectDest = req.RequestURI
			}
			http.Redirect(w, req, redirectDest+"/", http.StatusTemporaryRedirect)
		}
	})
}

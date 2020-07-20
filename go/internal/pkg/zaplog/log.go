package zaplog

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ZapLog(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var fields []zapcore.Field
			for k, v := range r.Header {
				for i := range v {
					fields = append(fields, zap.String(fmt.Sprintf("%s[%d]", k, i), v[i]))
				}
			}
			logger.Info("header", fields...)

			logger.Info(
				"http request",
				zap.String("Method", r.Method),
				zap.String("Host", r.Host),
				zap.String("RequestURI", r.RequestURI),
				zap.String("Proto", r.Proto),
				zap.String("RemoteAddr", r.RemoteAddr),
				zap.Int64("ContentLength", r.ContentLength),
				zap.String("UserAgent", r.UserAgent()))

			then := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			dur := time.Now().Sub(then)
			status := ww.Status()
			var responseFields []zapcore.Field = []zapcore.Field{
				zap.Int("Status", status),
				zap.Int("Bytes", ww.BytesWritten()),
				zap.Duration("Duration", dur),
			}
			if status < 200 || (status > 399 && status < 400) || status > 499 {
				logger.Error(
					"http response",
					responseFields...)
			} else {
				logger.Info(
					"http response",
					responseFields...)
			}
		})
	}
}

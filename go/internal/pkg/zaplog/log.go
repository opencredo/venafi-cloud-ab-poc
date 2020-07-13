package zaplog

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ZapLog(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			var fields []zapcore.Field = []zapcore.Field{
				zap.Int("Status", status),
				zap.Int("Bytes", ww.BytesWritten()),
				zap.Duration("Duration", dur),
			}
			if status < 200 || (status > 399 && status < 400) || status > 499 {
				logger.Error(
					"http response",
					fields...)
			} else {
				logger.Info(
					"http response",
					fields...)
			}
		})
	}
}

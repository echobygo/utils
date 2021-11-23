package MiaPrometheus

import (
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/kataras/iris/v12"
	"strconv"
)
//
//const (
//	httpClientReqsName    = "http_client_requests_total"
//	httpClientLatencyName = "http_client_duration_seconds"
//)

type prom interface {
	RegisterHistogram(histogram *prometheus.HistogramVec)
	RegisterCounter(conter *prometheus.CounterVec)
}

// NewClientPrometheus HTTP Client middleware that monitors requests made.
func NewClientPrometheus(serviceName string, p prom)( func(iris.Context)) {
	HttpServerCounterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        httpClientReqsName,
			Help:        "",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"domain", "http_code", "protocol", "method","path"},
	)
	p.RegisterCounter(HttpServerCounterVec)
	HttpServerTimerVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        httpClientLatencyName,
		Help:        "",
		ConstLabels: prometheus.Labels{"service": serviceName},
	},
		[]string{"domain", "http_code", "protocol", "method","path"},
	)
	p.RegisterHistogram(HttpServerTimerVec)

	return func(ctx iris.Context) {
		now := time.Now()
		ctx.Next()
		domain := ctx.Request().URL.Host
		method := ctx.Request().Method
		// rep := ctx.GetRespone()
		code := strconv.Itoa(ctx.GetStatusCode())
		protocol := ""
		// if rep.Error == nil {
		// 	protocol = rep.Proto
		// 	code = fmt.Sprint(rep.StatusCode)
		// } else {
		// 	code = fmt.Sprintf("dial tcp %s: i/o timeout", domain)
		// }
		HttpServerCounterVec.WithLabelValues(domain, code, protocol, method,ctx.Request().URL.Path).Inc()
		HttpServerTimerVec.WithLabelValues(domain, code, protocol, method,ctx.Request().URL.Path).Observe(float64(time.Since(now).Nanoseconds()) / 1000000000)
	}
}
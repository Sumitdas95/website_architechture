package httptrace

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	apmgo "github.com/deliveroo/apm-go"
)

// TraceAndServe See dd-trace-go/contrib/internal/httputil
func TraceAndServe(h http.Handler, w http.ResponseWriter, r *http.Request, resource string, apm apmgo.Service, statusErr bool, spanOptions ...apmgo.SpanOption) {
	// Create a new span.
	opts := []apmgo.SpanOption{apmgo.WithIncomingRequest(r)}
	opts = append(opts, spanOptions...)

	span := apm.NewSpan("http.request", resource, apmgo.SpanTypeWeb, opts...)
	span.SetTag("level", "INFO")
	span.SetTag("path", r.URL.Path)
	span.SetHTTPURL(r.URL.Path)
	span.SetHTTPMethod(r.Method)

	if clientID, _, ok := r.BasicAuth(); ok {
		span.SetTag("http.basicauth.clientID", clientID)
	}

	startTime := time.Now()
	httpResponseCode := http.StatusInternalServerError

	defer func() {
		duration := time.Since(startTime)

		// Track the distribution of http request durations in milliseconds while
		// ensuring to include the `http.status_code` if and when it is greater
		// than zero (request did not timeout).
		apm.StatsD().Distribution("roo.http.request.latency",
			float64(duration.Milliseconds()), 1, []string{
				"http.status_code", strconv.Itoa(httpResponseCode),
				"service", span.ServiceName(),
				"resource", span.Resource(),
			}...)

		if errFromRecover := recover(); errFromRecover != nil {
			w.WriteHeader(httpResponseCode)
			err := errorFromPanic(errFromRecover)
			span.FinishWithAlert(err)
			return
		}

		var err error
		if statusErr {
			rw, ok := w.(*TracingResponseWriter)
			if ok && (rw.StatusCode >= http.StatusInternalServerError || rw.StatusCode <= 0) {
				err = fmt.Errorf("%d: %s", rw.StatusCode, statusCodeText(rw.StatusCode))
			}
		}
		span.FinishWithError(err)
	}()

	// Embed span within context.
	ctx := apmgo.ContextWithSpan(r.Context(), span)
	r = r.WithContext(ctx)

	h.ServeHTTP(w, r)

	rw, ok := w.(*TracingResponseWriter)
	if ok && rw.StatusCode > 0 {
		httpResponseCode = rw.StatusCode
		span.SetTag("http.status_code", strconv.Itoa(httpResponseCode))
		span.SetTag("network.bytes_written", strconv.Itoa(rw.size))
	}
}

// TracingResponseWriter satisfies http.ResponseWriter interface, and allows
// capturing status code and bytes written to be used by the tracer middleware.
type TracingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	size       int
}

// NewTracingResponseWriter creates TracingResponseWriter wrapper
func NewTracingResponseWriter(w http.ResponseWriter) *TracingResponseWriter {
	return &TracingResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
}

// HTTPResponseWriter returns the wrapped ResponseWriter.
func (trw *TracingResponseWriter) HTTPResponseWriter() http.ResponseWriter {
	return trw.ResponseWriter
}

func (trw *TracingResponseWriter) WriteHeader(code int) {
	trw.StatusCode = code
	trw.ResponseWriter.WriteHeader(code)
}

func (trw *TracingResponseWriter) Write(b []byte) (int, error) {
	n, err := trw.ResponseWriter.Write(b)
	trw.size += n
	return n, err
}

// Hijack implements the http.Hijacker interface.
func (trw *TracingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if trw.size < 0 {
		trw.size = 0
	}
	return trw.ResponseWriter.(http.Hijacker).Hijack()
}

// Flush implements the http.Flusher interface.
func (trw *TracingResponseWriter) Flush() {
	trw.WriteHeaderNow()
	trw.ResponseWriter.(http.Flusher).Flush()
}

func (trw *TracingResponseWriter) WriteHeaderNow() {
	if !trw.Written() {
		trw.size = 0
		trw.ResponseWriter.WriteHeader(trw.StatusCode)
	}
}

func (trw *TracingResponseWriter) Written() bool {
	return trw.size != -1
}

// CloseNotify implements the http.CloseNotifier interface.
func (trw *TracingResponseWriter) CloseNotify() <-chan bool {
	return trw.ResponseWriter.(http.CloseNotifier).CloseNotify() //nolint:staticcheck
}

func (trw *TracingResponseWriter) Status() int {
	return trw.StatusCode
}

func (trw *TracingResponseWriter) Size() int {
	return trw.size
}

func (trw *TracingResponseWriter) Pusher() (pusher http.Pusher) {
	if pusher, ok := trw.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}

func (trw *TracingResponseWriter) WriteString(s string) (n int, err error) {
	trw.WriteHeaderNow()
	n, err = io.WriteString(trw.ResponseWriter, s)
	trw.size += n
	return
}

func statusCodeText(statusCode int) string {
	text := http.StatusText(statusCode)
	if text == "" {
		text = "non standard code"
	}
	return text
}

func errorFromPanic(err interface{}) error {
	var e error
	switch x := err.(type) {
	case string:
		e = fmt.Errorf("panic: %s", x)
	case error:
		e = fmt.Errorf("panic: %w", x)
	default:
		e = fmt.Errorf("panic: %#v", x)
	}
	return e
}

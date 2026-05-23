package observability

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
)

type ErrorInfo struct {
	Status            int
	Code              string
	GRPCCode          string
	DownstreamService string
}

type SpanInfo struct {
	TraceID string
	SpanID  string
}

type StatusRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int
	route       string
	errorInfo   ErrorInfo
	spanInfo    SpanInfo
	userIDHash  string
	wroteHeader bool
}

func NewStatusRecorder(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (r *StatusRecorder) Status() int {
	if r == nil {
		return http.StatusOK
	}
	return r.status
}

func (r *StatusRecorder) Bytes() int {
	if r == nil {
		return 0
	}
	return r.bytes
}

func (r *StatusRecorder) Route() string {
	if r == nil {
		return ""
	}
	return r.route
}

func (r *StatusRecorder) ErrorInfo() ErrorInfo {
	if r == nil {
		return ErrorInfo{}
	}
	return r.errorInfo
}

func (r *StatusRecorder) SpanInfo() SpanInfo {
	if r == nil {
		return SpanInfo{}
	}
	return r.spanInfo
}

func (r *StatusRecorder) UserIDHash() string {
	if r == nil {
		return ""
	}
	return r.userIDHash
}

func (r *StatusRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *StatusRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

func (r *StatusRecorder) ReadFrom(reader io.Reader) (int64, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	if rf, ok := r.ResponseWriter.(io.ReaderFrom); ok {
		n, err := rf.ReadFrom(reader)
		r.bytes += int(n)
		return n, err
	}
	n, err := io.Copy(r.ResponseWriter, reader)
	r.bytes += int(n)
	return n, err
}

func (r *StatusRecorder) Flush() {
	_ = http.NewResponseController(r.ResponseWriter).Flush()
}

func (r *StatusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return http.NewResponseController(r.ResponseWriter).Hijack()
}

func (r *StatusRecorder) Push(target string, opts *http.PushOptions) error {
	pusher, ok := r.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return pusher.Push(target, opts)
}

func (r *StatusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func RecordRoute(w http.ResponseWriter, route string) {
	for w != nil {
		if recorder, ok := w.(*StatusRecorder); ok {
			recorder.route = route
		}
		next, ok := w.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			return
		}
		w = next.Unwrap()
	}
}

func RecordError(w http.ResponseWriter, info ErrorInfo) {
	for w != nil {
		if recorder, ok := w.(*StatusRecorder); ok {
			recorder.errorInfo = info
		}
		next, ok := w.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			return
		}
		w = next.Unwrap()
	}
}

func RecordSpan(w http.ResponseWriter, info SpanInfo) {
	for w != nil {
		if recorder, ok := w.(*StatusRecorder); ok {
			recorder.spanInfo = info
		}
		next, ok := w.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			return
		}
		w = next.Unwrap()
	}
}

func RecordUserIDHash(w http.ResponseWriter, userIDHash string) {
	for w != nil {
		if recorder, ok := w.(*StatusRecorder); ok {
			recorder.userIDHash = userIDHash
		}
		next, ok := w.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			return
		}
		w = next.Unwrap()
	}
}

func RecorderFrom(w http.ResponseWriter) (*StatusRecorder, bool) {
	for w != nil {
		recorder, ok := w.(*StatusRecorder)
		if ok {
			return recorder, true
		}
		next, ok := w.(interface{ Unwrap() http.ResponseWriter })
		if !ok {
			return nil, false
		}
		w = next.Unwrap()
	}
	return nil, false
}

func IsClientClosed(status int) bool {
	return status == 499
}

var ErrRecorderUnavailable = errors.New("status recorder unavailable")

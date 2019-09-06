package httprint

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	requestLoggerKey = "requestLoggerKey"
	logLabel         = "log:"
	recycleLimit     = 1024 * 1024
)

var (
	Enable      bool   = false
	TimeFormat  string = "2006-01-02T15:04:05"
	PrintHeader bool   = false

	Output io.Writer
	outMu  sync.Mutex

	pool sync.Pool

	valueSeparator  = []byte{'\t'}
	lineSeparator   = []byte{'\n'}
	headerSeparator = []byte{','}
	labelSeparator  = []byte{':'}
)

func init() {
	Output = os.Stderr
}

type requestLogger struct {
	buf bytes.Buffer
	mu  sync.RWMutex
}

// リクエストが終わってからログをまとめて出力するので
// ここではバッファに入れるだけ
func (g *requestLogger) print(s string) {
	// 改行なしにしておく
	s = strings.ReplaceAll(s, "\n", " ")

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.buf.Len() > 0 {
		g.buf.WriteByte('\t')
	}

	g.buf.WriteString(logLabel)
	g.buf.WriteString(s)
}

//noinspection GoUnhandledErrorResult
func (g *requestLogger) dumpRequest(r *http.Request, start time.Time) {
	if g.empty() {
		return
	}

	duration := time.Now().Sub(start) / time.Millisecond
	reqTime := start.Format(TimeFormat)

	outMu.Lock()
	defer outMu.Unlock()

	fmt.Fprintf(Output, "time:%s\t"+
		"host:%s\t"+
		"method:%s\t"+
		"path:%s\t"+
		"reqtime:%dms",
		reqTime,
		r.RemoteAddr,
		r.Method,
		r.RequestURI,
		duration,
	)

	//noinspection GoBoolExpressions
	if PrintHeader {
		for hk, hvs := range r.Header {
			if len(hvs) <= 0 {
				continue
			}

			Output.Write(valueSeparator)
			io.WriteString(Output, hk)
			Output.Write(labelSeparator)

			for i, hv := range hvs {
				if i > 0 {
					Output.Write(headerSeparator)
				}
				io.WriteString(Output, hv)
			}
		}
	}

	g.dumpBuffer()
}

//noinspection GoUnhandledErrorResult
func (g *requestLogger) dumpBuffer() {
	g.mu.RLock()
	defer g.mu.RUnlock()

	Output.Write(valueSeparator)
	g.buf.WriteTo(Output)
	Output.Write(lineSeparator)
}

func (g *requestLogger) empty() bool {
	return g.buf.Len() <= 0
}

func (g *requestLogger) reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.buf.Reset()
}

func putRequestLogger(g *requestLogger) {
	g.reset()

	// あまりにも大きなロガーはプールに入れない
	if g.buf.Cap() > recycleLimit {
		return
	}

	pool.Put(g)
}

func getRequestLogger() *requestLogger {
	v := pool.Get()
	if v != nil {
		if g, ok := v.(*requestLogger); ok {
			return g
		}
	}
	return &requestLogger{}
}

func WrapHandler(handler http.HandlerFunc) http.HandlerFunc {
	if !Enable {
		return handler
	}

	return func(w http.ResponseWriter, r *http.Request) {
		g := getRequestLogger()
		defer putRequestLogger(g)

		// ロガーをcontextに入れる
		ctx := r.Context()
		ctx2 := context.WithValue(ctx, requestLoggerKey, g)
		r2 := r.WithContext(ctx2)

		// もとのハンドラを呼ぶ
		start := time.Now()
		handler(w, r2)

		// ログを出力する
		g.dumpRequest(r, start)
	}
}

func Print(r *http.Request, args ...interface{}) {
	logPrint(r, fmt.Sprint(args...))
}

func Printf(r *http.Request, format string, args ...interface{}) {
	logPrint(r, fmt.Sprintf(format, args...))
}

// リクエストが終わってからログをまとめて出力するので
// ここではバッファに入れるだけ
func logPrint(r *http.Request, s string) {
	ctx := r.Context()
	v := ctx.Value(requestLoggerKey)
	if v == nil {
		return
	}

	g, ok := v.(*requestLogger)
	if !ok || g == nil {
		return
	}

	g.print(s)
}

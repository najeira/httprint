package httprint

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
	"time"
)

func Test_empty(t *testing.T) {
	var buf bytes.Buffer
	Output = &buf

	const str = "this is test print"

	g := getRequestLogger()

	if !g.empty() {
		t.Error("not empty")
	}

	g.print(str)

	if g.empty() {
		t.Error("empty")
	}

	g.reset()

	if !g.empty() {
		t.Error("not empty")
	}
}

func Test_dumpBuffer(t *testing.T) {
	var buf bytes.Buffer
	Output = &buf

	const str = "this is test print"

	g := getRequestLogger()
	g.print(str)
	g.dumpBuffer()

	out := buf.String()
	if out != "\tlog:"+str+"\n" {
		t.Error(out)
	}
}

func Test_dumpRequest(t *testing.T) {
	var buf bytes.Buffer
	Output = &buf

	r, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	const str = "this is test print"

	g := getRequestLogger()
	g.print(str)
	g.dumpRequest(r, time.Time{})

	out := buf.String()
	if !strings.HasPrefix(out, "time:0001-01-01T00:00:00\thost:") {
		t.Error(out)
	}
	if !strings.HasSuffix(out, "\tlog:"+str+"\n") {
		t.Error(out)
	}
}

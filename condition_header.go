package conditionheader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type Config struct {
	Rules []*Rule `json:"rules,omitempty"`
}

type Rule struct {
	Conditions map[string]string `json:"conditions,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
}

func CreateConfig() *Config {
	return &Config{}
}

type ConditionHeader struct {
	next  http.Handler
	rules []*Rule
	name  string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.Rules) == 0 {
		return nil, fmt.Errorf("rules cannot be empty")
	}

	return &ConditionHeader{
		rules: config.Rules,
		next:  next,
		name:  name,
	}, nil
}

type wrappedResponseWriter struct {
	w    http.ResponseWriter
	buf  *bytes.Buffer
	code int
}

func (w *wrappedResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *wrappedResponseWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *wrappedResponseWriter) WriteHeader(code int) {
	w.code = code
}

func (w *wrappedResponseWriter) Flush() {
	w.w.WriteHeader(w.code)
	io.Copy(w.w, w.buf)
}

func (a *ConditionHeader) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	resp := &wrappedResponseWriter{
		w:    rw,
		buf:  &bytes.Buffer{},
		code: 200,
	}
	defer resp.Flush()

	a.next.ServeHTTP(resp, req)

	for _, rule := range a.rules {
		match := true
		for key, condition := range rule.Conditions {
			val := resp.Header().Get(key)
			if condition == "" {
				if val != "" {
					match = false
					break
				}
			} else {
				rex := regexp.MustCompile(condition)
				if val == "" || !rex.MatchString(val) {
					match = false
					break
				}
			}
		}

		if !match {
			return
		}

		for key, value := range rule.Headers {
			if value == "" {
				resp.Header().Del(key)
			} else {
				resp.Header().Set(key, value)
			}
		}
	}
}

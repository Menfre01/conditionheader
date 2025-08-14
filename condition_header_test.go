package conditionheader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConditionHeader(t *testing.T) {
	cfg := CreateConfig()
	cfg.Rules = append(cfg.Rules, Rule{
		Conditions: map[string]string{"Content-Type": "^text/html.*$"},
		Headers:    map[string]string{"Cache-Control": "no-cache, must-revalidate"},
	})

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/html")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("hello world"))
	})

	handler, err := New(ctx, next, cfg, "condition-header-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	assertHeader(t, recorder, "Cache-Control", "no-cache, must-revalidate")
}

func assertHeader(t *testing.T, writer http.ResponseWriter, key, expected string) {
	t.Helper()
	if writer.Header().Get(key) != expected {
		t.Errorf("invalid header value: %s", writer.Header().Get(key))
	}
}

func TestConditionHeader_ServeHTTP(t *testing.T) {
	type fields struct {
		next  http.Handler
		rules []Rule
		name  string
	}
	type args struct {
		rw  http.ResponseWriter
		req *http.Request
	}
	type want struct {
		headers map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "match",
			fields: fields{
				next: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					rw.Header().Set("Content-Type", "text/html; charset=utf-8")
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte("hello world"))
				}),
				rules: []Rule{
					{
						Conditions: map[string]string{"Content-Type": "text/html.*"},
						Headers:    map[string]string{"Cache-Control": "no-cache"},
					},
				},
				name: "condition-header-plugin",
			},
			args: args{
				rw:  httptest.NewRecorder(),
				req: httptest.NewRequest(http.MethodGet, "http://localhost", nil),
			},
			want: want{
				headers: map[string]string{"Cache-Control": "no-cache"},
			},
		},
		{
			name: "not match",
			fields: fields{
				next: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte("hello world"))
				}),
				rules: []Rule{
					{
						Conditions: map[string]string{"Content-Type": "text/plain.*"},
						Headers:    map[string]string{"Cache-Control": "no-cache"},
					},
				},
				name: "condition-header-plugin",
			},
			args: args{
				rw:  httptest.NewRecorder(),
				req: httptest.NewRequest(http.MethodGet, "http://localhost", nil),
			},
			want: want{
				headers: map[string]string{},
			},
		},
		{
			name: "empty condition",
			fields: fields{
				next: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte("hello world"))
				}),
				rules: []Rule{
					{
						Conditions: map[string]string{"Content-Type": ""},
						Headers:    map[string]string{"Content-Type": "text/plain; charset=utf-8"},
					},
				},
				name: "condition-header-plugin",
			},
			args: args{
				rw:  httptest.NewRecorder(),
				req: httptest.NewRequest(http.MethodGet, "http://localhost", nil),
			},
			want: want{
				headers: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &ConditionHeader{
				next:  tt.fields.next,
				rules: tt.fields.rules,
				name:  tt.fields.name,
			}
			a.ServeHTTP(tt.args.rw, tt.args.req)
			for key, value := range tt.want.headers {
				assertHeader(t, tt.args.rw, key, value)
			}
		})
	}
}

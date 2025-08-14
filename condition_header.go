package condition_header

import (
	"context"
	"fmt"
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
	return &Config{
		Rules: make([]*Rule, 0),
	}
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

func (a *ConditionHeader) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	a.next.ServeHTTP(rw, req)

	for _, rule := range a.rules {
		match := true
		for key, condition := range rule.Conditions {
			val := rw.Header().Get(key)
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
				rw.Header().Del(key)
			} else {
				rw.Header().Set(key, value)
			}
		}
	}
}

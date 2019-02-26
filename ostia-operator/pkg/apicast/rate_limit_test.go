package apicast

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
)

func TestProcessRateLimits(t *testing.T) {
	policyFromRlJson := func(rlCrd []byte) (PolicyChain, error) {
		var rl v1alpha1.RateLimit
		err := json.Unmarshal(rlCrd, &rl)
		if err != nil {
			t.Fatalf("error unmarshalling rate limit crd snippet to RateLimit struct - %s", err)
		}
		return processRateLimitPolicies([]v1alpha1.RateLimit{rl})
	}

	apiCastConfigFromPolicy := func(pc PolicyChain) []byte {
		conf, err := json.Marshal(pc)
		if err != nil {
			t.Fatalf("error unmarshalling apicast config")
		}
		return conf
	}

	fixedRateInputs := []struct {
		mockCrdDefinition []byte
		expectErr         bool
		expect            FixedWindowRateLimiter
		shouldContain     string
	}{
		{
			mockCrdDefinition: []byte(`{"type":"FixedWindow","name":"fixed","limit":"3600/hr"}`),
			expect: FixedWindowRateLimiter{
				Window: 3600,
				Count:  3600,
				Key:    LimiterKey{"fixed", "plain", "service"},
			},
			shouldContain: `{"fixed_window_limiters":[{"count":3600,"key":{"name":"fixed","name_type":"plain","scope":"service"},"window":3600}]}}`,
		},
		{
			mockCrdDefinition: []byte(`{"type":"FixedWindow","name":"testing_templating","limit":"10/s", "source":"{{remote_addr}}"}`),
			expect: FixedWindowRateLimiter{
				Window: 1,
				Count:  10,
				Key:    LimiterKey{"{{remote_addr}}", "liquid", "service"},
			},
			shouldContain: `{"fixed_window_limiters":[{"count":10,"key":{"name":"{{remote_addr}}","name_type":"liquid","scope":"service"},"window":1}]}}`,
		},
		{
			mockCrdDefinition: []byte(`{"type":"FixedWindow","name":"testing_default_time","limit":"100"}`),
			expect: FixedWindowRateLimiter{
				Window: 1,
				Count:  100,
				Key:    LimiterKey{"testing_default_time", "plain", "service"},
			},
			shouldContain: `{"fixed_window_limiters":[{"count":100,"key":{"name":"testing_default_time","name_type":"plain","scope":"service"},"window":1}]}}`,
		},
		{
			mockCrdDefinition: []byte(`{"type":"FixedWindow","name":"expect_err","limit":"ten"}`),
			expectErr:         true,
		},
	}
	for _, input := range fixedRateInputs {
		result, err := policyFromRlJson(input.mockCrdDefinition)
		if input.expectErr && err != nil {
			continue
		} else if err != nil {
			t.Errorf("unexpected error")
		}
		converted := *result.Configuration.FixedWindowLimiters
		equals(t, input.expect, converted[0])

		if !strings.Contains(string(apiCastConfigFromPolicy(result)), input.shouldContain) {
			t.Errorf("unexpected or missing config - \nshould contain - %s \nequals -%s",
				input.shouldContain, string(apiCastConfigFromPolicy(result)))
		}
	}

	leakyBucketInputs := []struct {
		mockCrdDefinition []byte
		expectErr         bool
		expect            LeakyBucketRateLimiter
		shouldContain     string
	}{
		{
			mockCrdDefinition: []byte(`{"type":"LeakyBucket","name":"fixed","limit":"100/m"}`),
			expect: LeakyBucketRateLimiter{
				Burst: 0,
				Key:   LimiterKey{"fixed", "plain", "service"},
				Rate:  1,
			},
			shouldContain: `{"leaky_bucket_limiters":[{"burst":0,"key":{"name":"fixed","name_type":"plain","scope":"service"},"rate":1}]}`,
		},
		{
			mockCrdDefinition: []byte(`{"type":"LeakyBucket","name":"fixed","limit":"120/m","burst":20}`),
			expect: LeakyBucketRateLimiter{
				Burst: 20,
				Key:   LimiterKey{"fixed", "plain", "service"},
				Rate:  2,
			},
			shouldContain: `{"leaky_bucket_limiters":[{"burst":20,"key":{"name":"fixed","name_type":"plain","scope":"service"},"rate":2}]}`,
		},
		{
			mockCrdDefinition: []byte(`{"type":"LeakyBucket","name":"fail","limit":"120xm","burst":20}`),
			expectErr:         true,
		},
		{
			mockCrdDefinition: []byte(`{"type":"LeakyBucket","name":"missing_limit","burst":20}`),
			expectErr:         true,
		},
	}
	for _, input := range leakyBucketInputs {
		result, err := policyFromRlJson(input.mockCrdDefinition)
		if input.expectErr && err != nil {
			continue
		} else if err != nil {
			t.Errorf("unexpected error")
		}
		converted := *result.Configuration.LeakyBucketLimiters
		equals(t, input.expect, converted[0])

		if !strings.Contains(string(apiCastConfigFromPolicy(result)), input.shouldContain) {
			t.Errorf("unexpected or missing config - \nshould contain - %s \nequals -%s",
				input.shouldContain, string(apiCastConfigFromPolicy(result)))
		}
	}

	connBasedInputs := []struct {
		mockCrdDefinition []byte
		expectErr         bool
		expect            ConnectionRateLimiter
		shouldContain     string
	}{
		{
			mockCrdDefinition: []byte(`{"type":"ConnectionBased","conn": 100, "name":"fixed","delay":10}`),
			expect: ConnectionRateLimiter{
				Burst: 0,
				Conn:  100,
				Delay: 10,
				Key:   LimiterKey{"fixed", "plain", "service"},
			},
			shouldContain: `{"connection_limiters":[{"burst":0,"conn":100,"delay":10,"key":{"name":"fixed","name_type":"plain","scope":"service"}}]}`,
		},
		{
			mockCrdDefinition: []byte(`{"type":"ConnectionBased","conn": 100, "name":"fixed","burst":10}`),
			expect: ConnectionRateLimiter{
				Burst: 10,
				Conn:  100,
				Delay: 0,
				Key:   LimiterKey{"fixed", "plain", "service"},
			},
			shouldContain: `{"connection_limiters":[{"burst":10,"conn":100,"delay":0,"key":{"name":"fixed","name_type":"plain","scope":"service"}}]}`,
		},
		{
			mockCrdDefinition: []byte(`{"type":"ConnectionBased", "name":"fixed","burst":10}`),
			expectErr:         true,
		},
	}

	for _, input := range connBasedInputs {
		result, err := policyFromRlJson(input.mockCrdDefinition)
		if input.expectErr && err != nil {
			continue
		} else if err != nil {
			t.Errorf("unexpected error")
		}
		converted := *result.Configuration.ConnectionLimiters
		equals(t, input.expect, converted[0])

		if !strings.Contains(string(apiCastConfigFromPolicy(result)), input.shouldContain) {
			t.Errorf("unexpected or missing config - \nshould contain - %s \nequals -%s",
				input.shouldContain, string(apiCastConfigFromPolicy(result)))
		}
	}

	someInvalidType := []byte(`{"type":"NA","conn": 100, "name":"fixed","delay":10}`)
	_, err := policyFromRlJson(someInvalidType)
	if err == nil {
		t.Fail()
	}
}

func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

package standalone

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRateLimitPolicyConfiguration(t *testing.T) {
	var policy = Policy{Name: "some", Configuration: RateLimitPolicyConfiguration{
		FixedWindowLimiters: &[]FixedWindowRateLimiter{
			{Key: LimiterKey{Name: "somekey", NameType: "plain", Scope: "service"},
				Window: 60, Count: 10}},
	}}
	var json, err = json.Marshal(policy)

	if err != nil {
		t.Errorf("marshal error %s", err)
	}

	if strings.Compare(string(json),
		`{"policy":"some","configuration":{"fixed_window_limiters":[{"count":10,"key":{"name":"somekey","name_type":"plain","scope":"service"},"window":60}]}}`) != 0 {
		t.Errorf("configuration is not correct: %s", string(json))
	}
}

func TestRateLimitPolicyConfigurationNull(t *testing.T) {
	var policy = Policy{Name: "some"}
	var json, err = json.Marshal(policy)

	if err != nil {
		t.Errorf("marshal error %s", err)
	}

	if strings.Compare(string(json),
		`{"policy":"some"}`) != 0 {
		t.Errorf("configuration is not correct: %s", string(json))
	}
}

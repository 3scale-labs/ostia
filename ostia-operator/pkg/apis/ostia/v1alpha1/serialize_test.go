package v1alpha1

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestUnmarshalCondition(t *testing.T) {
	inputs := []struct {
		crd       []byte
		expectC   *Condition
		expectErr bool
	}{
		{
			crd: []byte(`{"operation":"and", "operations":[{"http_method": "GET"}]}`),
			expectC: &Condition{
				Operator: "",
				Operations: []RateLimitCondition{
					MethodBasedCondition{Method: "GET"},
				},
			},
			expectErr: false,
		},
		{
			crd: []byte(`{"operation":"or", "operations":[{"http_method": "GET"}]}`),
			expectC: &Condition{
				Operator: "",
				Operations: []RateLimitCondition{
					MethodBasedCondition{Method: "GET"},
				},
			},
			expectErr: false,
		},
		{
			crd: []byte(`{"operation":"or", "operations":[{"http_method": "GET"},{"request_path":"/test", "op":"!="}]}`),
			expectC: &Condition{
				Operator: "or",
				Operations: []RateLimitCondition{
					MethodBasedCondition{Method: "GET"},
					PathBasedCondition{Path: "/test", Operation: "!="},
				},
			},
			expectErr: false,
		},
		{
			crd: []byte(`{"operation":"and", "operations":[{"http_method": "GET"},{"request_path":"/test", "op":"!="},{"header": "TEST", "value":"test"}]}`),
			expectC: &Condition{
				Operator: "and",
				Operations: []RateLimitCondition{
					MethodBasedCondition{Method: "GET"},
					PathBasedCondition{Path: "/test", Operation: "!="},
					HeaderBasedCondition{Header: "TEST", Value: "test"},
				},
			},
			expectErr: false,
		},
		{
			crd:       []byte(`{"operation":"and", "operations":[{"httP_method": "GET"}]}`),
			expectErr: true,
		},
		{
			crd:       []byte(`{"operations":[{"http_method": "GET"},{"request_path":"/test", "op":"!="}]}`),
			expectErr: true,
		},
	}
	for _, input := range inputs {
		c := &Condition{}
		err := c.UnmarshalJSON(input.crd)

		if input.expectErr && err != nil {
			continue
		} else if err != nil {
			t.Errorf("unexpected error")
		}

		if !reflect.DeepEqual(input.expectC, c) {
			fmt.Println(string(input.crd))
			t.Fatalf("unexpected condition parsed")
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	inputs := []struct {
		obj            RateLimitCondition
		expectContents string
		expectErr      bool
	}{
		{
			obj:            PathBasedCondition{Path: "/test_default_eq"},
			expectContents: `{"left":"{{uri}}","left_type":"liquid","op":"==","right":"/test_default_eq"}`,
		},
		{
			obj:            PathBasedCondition{Path: "/test_default_eq", Operation: "!="},
			expectContents: `{"left":"{{uri}}","left_type":"liquid","op":"!=","right":"/test_default_eq"}`,
		},
		{
			obj:       PathBasedCondition{Path: "/test_default_eq", Operation: "invalid"},
			expectErr: true,
		},
		{
			obj:       PathBasedCondition{Path: "/invalid*", Operation: "invalid"},
			expectErr: true,
		},
		{
			obj:       PathBasedCondition{Path: "\bad_path*"},
			expectErr: true,
		},
		{
			obj:            HeaderBasedCondition{Header: "test", Value: "example"},
			expectContents: `{"left":"{{headers['test']}}","left_type":"liquid","op":"==","right":"example"}`,
		},
		{
			obj:       HeaderBasedCondition{Header: "", Value: "example"},
			expectErr: true,
		},
		{
			obj:       HeaderBasedCondition{Header: "test", Value: ""},
			expectErr: true,
		},
		{
			obj:            MethodBasedCondition{Method: "GET"},
			expectContents: `{"left":"{{http_method}}","left_type":"liquid","op":"==","right":"GET"}`,
		},
		{
			obj:            MethodBasedCondition{Method: "get"},
			expectContents: `{"left":"{{http_method}}","left_type":"liquid","op":"==","right":"GET"}`,
		},
		{
			obj:       MethodBasedCondition{Method: "INVALID"},
			expectErr: true,
		},
	}

	for _, input := range inputs {
		res, err := input.obj.MarshalJSON()
		if input.expectErr && err != nil {
			continue
		} else if err != nil {
			t.Errorf("unexpected error")
		}
		if !strings.Contains(string(res), input.expectContents) {
			t.Errorf("unexpected or missing config - \nshould contain - %s \nequals -%s",
				input.expectContents, string(res))
		}
	}
}

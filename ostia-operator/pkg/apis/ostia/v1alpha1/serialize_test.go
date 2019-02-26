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
					methodBasedCondition{Method: "GET"},
				},
			},
			expectErr: false,
		},
		{
			crd: []byte(`{"operation":"or", "operations":[{"http_method": "GET"}]}`),
			expectC: &Condition{
				Operator: "",
				Operations: []RateLimitCondition{
					methodBasedCondition{Method: "GET"},
				},
			},
			expectErr: false,
		},
		{
			crd: []byte(`{"operation":"or", "operations":[{"http_method": "GET"},{"request_path":"/test", "op":"!="}]}`),
			expectC: &Condition{
				Operator: "or",
				Operations: []RateLimitCondition{
					methodBasedCondition{Method: "GET"},
					pathBasedCondition{Path: "/test", Operation: "!="},
				},
			},
			expectErr: false,
		},
		{
			crd: []byte(`{"operation":"and", "operations":[{"http_method": "GET"},{"request_path":"/test", "op":"!="},{"header": "TEST", "value":"test"}]}`),
			expectC: &Condition{
				Operator: "and",
				Operations: []RateLimitCondition{
					methodBasedCondition{Method: "GET"},
					pathBasedCondition{Path: "/test", Operation: "!="},
					headerBasedCondition{Header: "TEST", Value: "test"},
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
			obj:            pathBasedCondition{Path: "/test_default_eq"},
			expectContents: `{"left":"{{uri}}","left_type":"liquid","op":"==","right":"/test_default_eq"}`,
		},
		{
			obj:            pathBasedCondition{Path: "/test_default_eq", Operation: "!="},
			expectContents: `{"left":"{{uri}}","left_type":"liquid","op":"!=","right":"/test_default_eq"}`,
		},
		{
			obj:       pathBasedCondition{Path: "/test_default_eq", Operation: "invalid"},
			expectErr: true,
		},
		{
			obj:       pathBasedCondition{Path: "/invalid*", Operation: "invalid"},
			expectErr: true,
		},
		{
			obj:       pathBasedCondition{Path: "\bad_path*"},
			expectErr: true,
		},
		{
			obj:            headerBasedCondition{Header: "test", Value: "example"},
			expectContents: `{"left":"{{headers['test']}}","left_type":"liquid","op":"==","right":"example"}`,
		},
		{
			obj:       headerBasedCondition{Header: "", Value: "example"},
			expectErr: true,
		},
		{
			obj:       headerBasedCondition{Header: "test", Value: ""},
			expectErr: true,
		},
		{
			obj:            methodBasedCondition{Method: "GET"},
			expectContents: `{"left":"{{http_method}}","left_type":"liquid","op":"==","right":"GET"}`,
		},
		{
			obj:            methodBasedCondition{Method: "get"},
			expectContents: `{"left":"{{http_method}}","left_type":"liquid","op":"==","right":"GET"}`,
		},
		{
			obj:       methodBasedCondition{Method: "INVALID"},
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

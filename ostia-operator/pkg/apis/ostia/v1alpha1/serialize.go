package v1alpha1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (hc HeaderBasedCondition) MarshalJSON() ([]byte, error) {
	var op string
	if hc.Header == "" || hc.Value == "" {
		return nil, errors.New("header and header value required for header based condition")
	}

	op, err := parseOp(hc.Operation)
	if err != nil {
		return nil, err
	}

	condition := make(map[string]string)
	condition["left"] = fmt.Sprintf("{{headers['%s']}}", hc.Header)
	condition["left_type"] = "liquid"
	condition["op"] = op
	condition["right"] = hc.Value

	b, err := json.Marshal(condition)
	if err != nil {
		return nil, errors.New("error marshalling condition")
	}

	return b, nil
}

func (mc MethodBasedCondition) MarshalJSON() ([]byte, error) {
	op, err := parseOp(mc.Operation)
	if err != nil {
		return nil, err
	}

	method := strings.ToUpper(mc.Method)
	switch method {
	case http.MethodGet, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodPut:
		break
	default:
		return nil, errors.New("unsupported http method")

	}

	condition := make(map[string]string)
	condition["left"] = "{{http_method}}"
	condition["left_type"] = "liquid"
	condition["op"] = op
	condition["right"] = method

	b, err := json.Marshal(condition)
	if err != nil {
		return nil, errors.New("error marshalling condition")
	}

	return b, nil
}

func (pc PathBasedCondition) MarshalJSON() ([]byte, error) {
	var op string
	condition := make(map[string]string)

	if _, err := url.Parse("http://valid.com" + pc.Path); err != nil || pc.Path == "" {
		return nil, errors.New("invalid path provided")
	}

	op, err := parseOp(pc.Operation)
	if err != nil {
		return nil, err
	}

	condition["left"] = "{{uri}}"
	condition["left_type"] = "liquid"
	condition["op"] = op
	condition["right"] = pc.Path

	b, err := json.Marshal(condition)
	if err != nil {
		return nil, errors.New("error marshalling condition")
	}

	return b, nil
}

// UnmarshalJSON is responsible for converting rate limits read via CRD
// to concrete types which implement the RateLimitCondition interface
func (c *Condition) UnmarshalJSON(b []byte) error {
	var raw map[string]json.RawMessage
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return fmt.Errorf("error parsing rate limit conditions - %s", err)
	}

	var operationsErr error
	if operations, ok := raw["operations"]; ok {
		operationsErr = c.setOperations(operations)
	} else {
		operationsErr = errors.New("operations required for rate limit conditions")
	}

	if operationsErr != nil {
		return operationsErr
	}

	if len(c.Operations) > 1 {
		operationErr := errors.New("conditional operator must be set")
		if operator, ok := raw["operation"]; ok {
			operationErr = json.Unmarshal(operator, &c.Operator)

		}
		if operationsErr != nil || c.Operator == "" {
			return operationErr
		}

	}

	return nil

}

func (c *Condition) setOperations(b json.RawMessage) error {
	var operations []json.RawMessage

	err := json.Unmarshal(b, &operations)
	if err != nil {
		return err
	}

	for _, op := range operations {
		var condition map[string]json.RawMessage
		err := json.Unmarshal(op, &condition)
		if err != nil {
			return err
		}

		conditionErr := errors.New("unknown condition")

		if _, ok := condition["header"]; ok {
			var h HeaderBasedCondition
			conditionErr = json.Unmarshal(op, &h)
			// FIXME: c.Operations = append(c.Operations, h)

		} else if _, ok := condition["http_method"]; ok {
			var hm MethodBasedCondition
			conditionErr = json.Unmarshal(op, &hm)
			// FIXME: c.Operations = append(c.Operations, hm)

		} else if _, ok := condition["request_path"]; ok {
			var rp PathBasedCondition
			conditionErr = json.Unmarshal(op, &rp)
			// FIXME: c.Operations = append(c.Operations, rp)

		}

		if conditionErr != nil {
			return fmt.Errorf("error calling unmarshal - %s ", conditionErr)
		}

	}
	return nil

}

func parseOp(op string) (string, error) {
	if op != "" {
		switch op {
		case "==", "!=":
			break
		default:
			return "", errors.New("unrecognised operand provided")
		}
	} else {
		op = "=="
	}
	return op, nil
}

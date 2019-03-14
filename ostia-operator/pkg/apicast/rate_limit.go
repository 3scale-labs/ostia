package apicast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
)

const rateLimitPolicyName = "apicast.policy.rate_limit"

func processRateLimitPolicies(limits []v1alpha1.RateLimit) (PolicyChain, error) {
	var pc PolicyChain
	var fixedLimiters []FixedWindowRateLimiter
	var leakyLimiters []LeakyBucketRateLimiter
	var connLimiters []ConnectionRateLimiter

	for _, limit := range limits {
		switch limiterType := limit.Type; limiterType {
		case "FixedWindow":
			fw, err := toFixedWindow(limit)
			if err != nil {
				return pc, err
			}
			fixedLimiters = append(fixedLimiters, fw)
		case "LeakyBucket":
			lb, err := toLeakyBucket(limit)
			if err != nil {
				return pc, err
			}
			leakyLimiters = append(leakyLimiters, lb)
		case "ConnectionBased":
			cb, err := toConnectionBased(limit)
			if err != nil {
				return pc, err
			}
			connLimiters = append(connLimiters, cb)
		default:
			return pc, fmt.Errorf("missing or unknown 'type' field on %s rate limit definition", limit.Name)
		}
	}

	var config PolicyChainConfiguration
	if len(fixedLimiters) > 0 {
		config.FixedWindowLimiters = &fixedLimiters
	}
	if len(leakyLimiters) > 0 {
		config.LeakyBucketLimiters = &leakyLimiters
	}
	if len(connLimiters) > 0 {
		config.ConnectionLimiters = &connLimiters
	}

	pc.Name = rateLimitPolicyName
	pc.Configuration = config
	return pc, nil
}

func toFixedWindow(rl v1alpha1.RateLimit) (FixedWindowRateLimiter, error) {
	count, window, err := parseTimeLimits(rl)
	if err != nil {
		return FixedWindowRateLimiter{}, err
	}

	fw := FixedWindowRateLimiter{
		Condition: rl.Conditions,
		Count:     count,
		Key:       parseLimiterKey(rl),
		Window:    window,
	}

	return fw, nil
}

func toLeakyBucket(rl v1alpha1.RateLimit) (LeakyBucketRateLimiter, error) {
	var burst int

	rate, seconds, err := parseTimeLimits(rl)
	if err != nil {
		return LeakyBucketRateLimiter{}, err
	}

	if rl.Burst == nil || *rl.Burst < 0 {
		log.Info("setting 'burst' value for %s to 0", rl.Name)
	} else {
		burst = *rl.Burst
	}

	return LeakyBucketRateLimiter{burst, rl.Conditions, parseLimiterKey(rl), rate / seconds}, nil
}

func toConnectionBased(rl v1alpha1.RateLimit) (ConnectionRateLimiter, error) {
	var burst, conn, delay int

	if rl.Conn == nil || *rl.Conn < 1 {
		return ConnectionRateLimiter{}, fmt.Errorf("required property 'conn' not valid for rate limit %s", rl.Limit)
	}
	conn = *rl.Conn

	if rl.Burst == nil || *rl.Burst < 0 {
		log.Info("setting 'burst' value for %s to 0", rl.Name)
	} else {
		burst = *rl.Burst
	}

	if rl.Delay == nil || *rl.Delay < 0 {
		log.Info("setting 'delay' value for %s to 0", rl.Name)
	} else {
		delay = *rl.Delay
	}

	return ConnectionRateLimiter{burst, rl.Conditions, conn, delay, parseLimiterKey(rl)}, nil
}

func parseTimeLimits(rl v1alpha1.RateLimit) (int, int, error) {
	var requests, seconds int
	if rl.Limit == "" {
		return requests, seconds, fmt.Errorf("required property 'limit' missing from rate limit %s", rl.Limit)
	}
	seconds = 1
	parsedLimitVal := strings.Split(rl.Limit, "/")
	requests, err := strconv.Atoi(parsedLimitVal[0])
	if err != nil || requests < 1 {
		return requests, seconds, fmt.Errorf("'limit' value  for %s must be a non-negative integer", rl.Limit)
	}

	if len(parsedLimitVal) == 2 {
		switch parsedLimitVal[1] {
		case "s":
			break
		case "m":
			seconds = 60
		case "hr":
			seconds = 60 * 60
		default:
			fmt.Printf("unrecognised unit of time %s, for rate limit %s. defaulting to seconds", parsedLimitVal[1], rl.Limit)
		}
	}
	return requests, seconds, nil
}

func parseLimiterKey(rl v1alpha1.RateLimit) LimiterKey {
	key := LimiterKey{rl.Name, "plain", "service"}
	if rl.Source != "" {
		key.Name = rl.Source
		key.NameType = "liquid"
	}
	return key
}

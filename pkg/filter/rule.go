package filter

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/server"
	"reflect"
	"regexp"
	"strings"
)

func ParseRuleSet(ruleSet RuleSet, headers server.MessageHeaders) (bool, error) {
	var err error

	for _, rule := range ruleSet {
		matched, err := parseRuleAgainstHeaders(rule, headers)
		if err != nil {
			return false, err
		}

		if matched {
			return true, err
		}
	}

	return false, err
}

func parseRuleAgainstHeaders(rule Rule, headers server.MessageHeaders) (bool, error) {
	var err error

	for op, patterns := range rule {
		op = strings.ToLower(op)

		switch op {
		case "or":
			for _, pattern := range patterns {
				for patternHeaderName, patternValues := range pattern {
					patternHeaderName := strings.ToLower(patternHeaderName)

					if _, keyInMap := headers[patternHeaderName]; !keyInMap {
						continue
					}

					if matched, err := checkRulePattern(patternValues, headers[patternHeaderName]); err != nil {
						return false, err
					} else if matched {
						return true, nil
					}
				}
			}
		case "and":
			var patternMatched bool

			for _, pattern := range patterns {
				for patternHeaderName, patternValues := range pattern {
					patternHeaderName := strings.ToLower(patternHeaderName)

					if _, keyInMap := headers[patternHeaderName]; !keyInMap {
						return false, nil
					}

					if matched, err := checkRulePattern(patternValues, headers[patternHeaderName]); err != nil {
						return false, err
					} else if !matched {
						return false, nil
					} else if matched {
						patternMatched = true
					}
				}
			}

			return patternMatched, nil
		default:
			return false, fmt.Errorf("rule operator %q is unsupported", op)
		}
	}

	return false, err
}

func checkRulePattern(patternValues interface{}, headers interface{}) (bool, error) {
	parsedValues, err := parsePatternValues(patternValues)
	if err != nil {
		return false, err
	}

	for _, patternValue := range parsedValues {
		var headerList []string

		switch h := headers.(type) {
		case string:
			headerList = append(headerList, h)
		case []string:
			headerList = h
		default:
			return false, fmt.Errorf("unsupported header type %q", reflect.TypeOf(headers))
		}

		for _, header := range headerList {
			matched, err := checkMatch(patternValue, header)
			if err != nil {
				return false, err
			}

			if matched {
				return true, nil
			}
		}
	}

	return false, nil
}

func checkMatch(pattern string, s string) (bool, error) {
	patternLowered := strings.ToLower(pattern)
	s = strings.ToLower(s)
	var err error

	//fmt.Printf("%q == %q\n", pattern, s)

	if pattern == "" && s == "" {
		return true, err
	}

	if pattern == "" && s != "" {
		return false, err
	}

	if patternLowered == s {
		return true, err
	}

	if strings.Contains(s, patternLowered) {
		return true, err
	}

	regEx, err := regexp.Compile(fmt.Sprintf("(?i)%v", pattern))
	if err != nil {
		return false, err
	}

	if regEx.MatchString(s) {
		return true, err
	}

	return false, err
}

func parsePatternValues(patternValues interface{}) ([]string, error) {
	var values []string

	switch v := patternValues.(type) {
	case string:
		return append(values, v), nil
	case int:
		return append(values, fmt.Sprintf("%v", v)), nil
	//case float32:
	//	return append(values, fmt.Sprintf("%v", v)), nil
	//case float64:
	//	return append(values, fmt.Sprintf("%v", v)), nil
	//case bool:
	//	return append(values, fmt.Sprintf("%v", v)), nil
	case []string:
		for _, val := range v {
			values = append(values, val)
		}
	case []interface{}:
		for _, val := range v {
			p, err := parsePatternValues(val)

			if err != nil {
				return values, err
			}

			values = append(values, p...)
		}
	default:
		return values, fmt.Errorf("unsupported value type %v", reflect.TypeOf(v))
	}

	return values, nil
}

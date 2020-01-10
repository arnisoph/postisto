package filter_test

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/filter"
	"github.com/arnisoph/postisto/pkg/server"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

func TestParseRuleSet(t *testing.T) {
	require := require.New(t)

	// ACTUAL TESTS BELOW

	ruleParserTests := []struct {
		filters       map[string]filter.Filter
		matchExpected bool
		err           string
	}{
		{
			filters: map[string]filter.Filter{
				"simple 1o1 comparison": {
					RuleSet: filter.RuleSet{
						{
							"or": []map[string]interface{}{
								{"from": "foo@example.com"},
							},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: map[string]filter.Filter{
				"simple 101 comparison in or": {
					RuleSet: filter.RuleSet{
						{
							"or": []map[string]interface{}{
								{"from": "oO"},
								{"from": "foo@example.com"},
							},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: map[string]filter.Filter{
				"failing simple comparison": {
					RuleSet: filter.RuleSet{
						{
							"or": []map[string]interface{}{{"from": "wrong value"}},
						},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: map[string]filter.Filter{
				"comparison with uppercase text": {
					RuleSet: filter.RuleSet{
						{
							"and": []map[string]interface{}{
								{"from": "foo@example.com"},
								{"to": "me@EXAMPLE.com"},
							},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: map[string]filter.Filter{
				"failing and comparison": {
					RuleSet: filter.RuleSet{
						{
							"and": []map[string]interface{}{
								{"from": "you"},
								{"to": "you"},
							},
						},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: map[string]filter.Filter{
				"failing or comparison": {
					RuleSet: filter.RuleSet{
						{
							"or": []map[string]interface{}{
								{"from": "you"},
								{"to": "you"},
							},
						},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: map[string]filter.Filter{
				"failing with unsupported op": {
					RuleSet: filter.RuleSet{
						{
							"non-existent-op": []map[string]interface{}{
								{"from": "you"},
								{"to": "you"},
							},
						},
					},
				},
			},
			matchExpected: false,
			err:           `rule operator "non-existent-op" is unsupported`,
		},
		{
			filters: map[string]filter.Filter{
				"invalid value type": {
					RuleSet: filter.RuleSet{
						{
							"or": []map[string]interface{}{
								{"from": server.Connection{}},
							},
						},
					},
				},
			},
			matchExpected: false,
			err:           `unsupported value type server.Connection`,
		},
		{
			filters: map[string]filter.Filter{
				"invalid nested value type": {
					RuleSet: filter.RuleSet{
						{
							"or": []map[string]interface{}{
								{"from": []interface{}{"wrong1", "wrong2", "42", []interface{}{server.Connection{}}}},
							},
						},
					},
				},
			},
			matchExpected: false,
			err:           `unsupported value type server.Connection`,
		},
		{
			filters: map[string]filter.Filter{
				"substring comparison with and": {
					RuleSet: filter.RuleSet{
						{
							"and": []map[string]interface{}{{"from": "@example.com"}},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: map[string]filter.Filter{
				"substring comparison with or": {
					RuleSet: filter.RuleSet{
						{
							"or": []map[string]interface{}{{"from": "@example.com"}},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: map[string]filter.Filter{
				"failing on search for empty header": {
					RuleSet: filter.RuleSet{
						{
							"and": []map[string]interface{}{{"from": ""}},
						},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: map[string]filter.Filter{
				"successfully searching for empty header": {
					RuleSet: filter.RuleSet{
						{
							"and": []map[string]interface{}{{"empty-header": ""}},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: map[string]filter.Filter{
				"testing with ütf-8": {
					RuleSet: filter.RuleSet{
						{
							"and": []map[string]interface{}{{"subject": "löv"}},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: map[string]filter.Filter{
				"uppercase in rule + substring comparison": {
					RuleSet: filter.RuleSet{
						{
							"and": []map[string]interface{}{{"from": "@EXAMPLE.COM"}},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: map[string]filter.Filter{
				"uppercase in header comparison": {
					RuleSet: filter.RuleSet{
						{
							"and": []map[string]interface{}{{"to": "@example.com"}},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: map[string]filter.Filter{
				"regex comparison": {
					RuleSet: filter.RuleSet{
						{
							"and": []map[string]interface{}{
								{"subject": "löve$"},
								{"subject": "^with löve$"},
								{"subject": "^wit.*ve$"},
								{"subject": "^with\\s+löve$"},
								{"subject": "^.*$"},
								{"subject": ".*"},
								{"subject": "^with\\s+l(ö|ä)ve$"},
								{"suBject": "^with\\s+l(?:ö|ä)ve$"},
								{"subject": "^WITH"},
							},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: map[string]filter.Filter{
				"comparison with bad regex (and)": {
					RuleSet: filter.RuleSet{
						{
							"and": []map[string]interface{}{{"tO": "!^\\ü^@example.com"}},
						},
					},
				},
			},
			matchExpected: false,
			err:           "error parsing regexp: invalid escape sequence: `\\ü`",
		},
		{
			filters: map[string]filter.Filter{
				"comparison with bad regex (or)": {
					RuleSet: filter.RuleSet{
						{
							"or": []map[string]interface{}{{"to": "!^\\ü^@example.com"}},
						},
					},
				},
			},
			matchExpected: false,
			err:           "error parsing regexp: invalid escape sequence: `\\ü`",
		},
		{
			filters: map[string]filter.Filter{
				"several rules in ruleSet success": {
					RuleSet: filter.RuleSet{
						{"and": []map[string]interface{}{{"to": "@example.com"}}},
						{"or": []map[string]interface{}{{"subject": "löv"}}},
						{"and": []map[string]interface{}{{"from": ""}}},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: map[string]filter.Filter{
				"several rules in ruleSet failing": {
					RuleSet: filter.RuleSet{
						{"and": []map[string]interface{}{{"to": "@examplde.com"}}},
						{"or": []map[string]interface{}{{"sUbject": "löasdv"}}},
						{"and": []map[string]interface{}{{"from": ""}}},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: map[string]filter.Filter{
				"1o1 comparison with multiple values": {
					RuleSet: filter.RuleSet{
						{
							"or": []map[string]interface{}{
								{"froM": []string{"foo@example.com", "example.com", "foo"}},
							},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: map[string]filter.Filter{
				"101 comparison in or with multiple values": {
					RuleSet: filter.RuleSet{
						{
							"or": []map[string]interface{}{
								{"froM": "oO"},
								{"from": []string{"foo@example.com", "example.com", "foo"}},
							},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: map[string]filter.Filter{
				"101 comparison in OR with multiple values (failing)": {
					RuleSet: filter.RuleSet{
						{
							"or": []map[string]interface{}{
								{"From": "baz"},
								{"from": []interface{}{"wrong1", "wrong2", "42", 42}},
							},
						},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: map[string]filter.Filter{
				"101 comparison in AND with multiple values (failing)": {
					RuleSet: filter.RuleSet{
						{
							"and": []map[string]interface{}{
								{"from": "baz"},
								{"from": []string{"foo@example.com", "example.com", "foo"}},
							},
						},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: map[string]filter.Filter{
				"weirdest bug so far": {
					RuleSet: filter.RuleSet{
						{
							"and": []map[string]interface{}{
								{"X-Custom-Mail-Id": "16"},
								{"X-Notes-Item": "CSMemoFrom"},
							},
						},
					},
				},
			},
			matchExpected: false,
		},

		//{
		//	headers: MailHeaders{"from": "oO"},
		//	rule: filter.Rule{
		//		"or": []map[string]interface{}{
		//			{
		//				"or": filter.Rule{
		//					"or": []map[string]interface{}{
		//						{"from": "nope"},
		//						{"from": "oO"},
		//					},
		//				},
		//			},
		//		},
		//	},
		//	matchExpected: true,
		//},
	}

	testMailHeaders := server.MessageHeaders{"from": "foo@example.com", "to": "me@EXAMPLE.com", "subject": "With Löve", "empty-header": "", "custom-Header": "Foobar"}

	cfg, err := config.NewConfigFromFile("../../test/data/configs/valid/test/TestParserRuleSet.yaml")
	require.NoError(err)

	acc := cfg.Accounts["test"]
	filters := cfg.Filters["test"]
	require.NotNil(acc)

	for i, test := range ruleParserTests {
		for filterName, testFilter := range test.filters {
			// Test with native synthetic test data
			matched, err := filter.ParseRuleSet(testFilter.RuleSet, testMailHeaders)
			if test.err == "" {
				require.NoError(err)
			}
			if test.err != "" && err != nil {
				require.True(strings.HasPrefix(err.Error(), test.err), "NATIVE DATA TEST: Actual error message: %v", err.Error())
			}

			require.Equal(test.matchExpected, matched, "NATIVE DATA TEST: Test #%v (%q) from ruleParserTests failed! ruleSet=%q testMailHeaders=%q", i+1, filterName, testFilter.RuleSet, testMailHeaders)

			if filterName == "invalid value type" || filterName == "invalid nested value type" {
				// can't test NON-JSON data types in YAML
				continue
			}

			// Test with same synthetic ruleSet test data from YAML
			yml, err := yaml.Marshal(test.filters)
			_, fieldInMap := filters[filterName]
			require.True(fieldInMap, "Add test %q to TestParserRuleSet.yml:\n=========\n%v=========\n%v", filterName, string(yml), err)

			ymlFilter := filters[filterName]
			require.NotNil(ymlFilter)
			require.NotNil(ymlFilter.RuleSet)

			matched, err = filter.ParseRuleSet(ymlFilter.RuleSet, testMailHeaders)
			if test.err == "" {
				require.NoError(err)
			}
			if test.err != "" && err != nil {
				require.True(strings.HasPrefix(err.Error(), test.err), "YML DATA TEST: Actual error message: %v", err.Error())
			}
			require.Equal(test.matchExpected, matched, "YML DATA TEST: Test #%v (%q) from ruleParserTests failed: ruleSet=%q testMailHeaders=%q", i+1, filterName, ymlFilter.RuleSet, testMailHeaders)
		}
	}
}

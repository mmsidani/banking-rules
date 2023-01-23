package core

import (
	"encoding/json"
	"strings"

	"github.com/Knetic/govaluate"

	c "../common"
)

// SortConflicts takes group rules and individual rules, for example, and overrule the group rule when conflict is detected. In general, rules in second array overrule those in first. Problem: how to detect conflicting rules? one idea: rules concerning the exact same quantities and only those (for example, initiator and amount and nothing else) should be solved for conflicts (using SAT?)
func SortConflicts(m map[string][]ARule) []ARule {
	// TODO TODO. z3 (?) or 2 rules that use exact same parameters or maybe one's parameters is a subset of the other. rules in "spec" (specific) override those in "gen" (generic). so map m expected to have two keys

	return append(m["gen"], m["spec"]...) // so just concatenate until we have a proper implementation
}

// ARule structure for storing rules. rules are required to be ternary expressions. Rule must be a ternary expression that returns output from NofM() in RuleFunctions() or "nil"
type ARule struct {
	Rule     string `json:"rule"`
	RuleHash string `json:"rulehash"`
}

// Evaluate the rule for the given parameters. Currently returns "nil" (yes, string) or output from rule function(s)
func (r *ARule) Evaluate(m map[string]interface{}) interface{} {
	rule, err := govaluate.NewEvaluableExpressionWithFunctions(r.Rule, RuleFunctions())
	if err != nil {
		panic(err)
	}
	result, err := rule.Evaluate(m)
	if err != nil {
		panic(err)
	}

	return result
}

// NewRule constructs new rule
func NewRule(r string, ruleHash string) ARule {
	ret := ARule{Rule: r, RuleHash: ruleHash}

	// TODO checking if parameters are valid. costly?
	rule, err := govaluate.NewEvaluableExpressionWithFunctions(r, RuleFunctions())
	if err != nil {
		panic(err)
	}
	expparams := rule.Vars()
	for _, v := range expparams {
		if !c.RuleVariablesSet[v] {
			panic("wrong variable")
		}
	}

	// necessary (but not sufficient) condition for r to be ternary and for it to return "nil" when predicate is false and therefore no action is required
	if !strings.Contains(r, ":") || !strings.Contains(r, "?") || !strings.Contains(r, "nil") {
		panic("rule must be a ternary expression that returns 'nil' (yes, string) when predicate is false")
	}

	return ret
}

func unmarshalRules(rules [][]byte) (accountRules []ARule) {
	accountRules = make([]ARule, len(rules))
	for i, rule := range rules {
		var t ARule
		err := json.Unmarshal(rule, &t)
		if err != nil {
			panic(err)
		}
		accountRules[i] = t
	}

	return
}

func tabulateRules(rules []ARule) (ruleHashes []string, accountRules []string) {
	ruleHashes = make([]string, len(rules))
	accountRules = make([]string, len(rules))
	for i, rule := range rules {
		ruleHashes[i] = rule.RuleHash
		accountRules[i] = rule.Rule
	}

	return
}

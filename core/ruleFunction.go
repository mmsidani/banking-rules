package core

import (
	"github.com/Knetic/govaluate"
)

// helper function needed below.
func intersect(a, b []string) (c []string) {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			c = append(c, item)
		}
	}
	return
}

// rather than actually find the intersection, this one just says true/false as to whether the 2 arrays intersect
func fastIntersect(a, b []string) bool {
	m := make(map[string]bool)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			return true
		}
	}
	return false
}

func aggregateSpend(args ...interface{}) (interface{}, error) {
	// go to core banking system to get history of transactions and aggregate
	// to be added to RuleFunctions() output after implementation

	return nil, nil
}

func accountBalance(args ...interface{}) (interface{}, error) {
	// get the account balance from the CBS
	// to be added to RuleFunctions() output after implementation

	return nil, nil
}

// RuleFunctions the functions allowable in rule expressions
func RuleFunctions() map[string]govaluate.ExpressionFunction {

	ruleFunctions := map[string]govaluate.ExpressionFunction{
		"NofM": func(args ...interface{}) (interface{}, error) {

			a := args[0] // min number of signatures, i.e., N
			b := args[1] // string with comma-separated list of all authorised signers by ID (i.e., initiator), M is the len() of the list

			return []interface{}{a, b}, nil
		},
	}

	// now add the regular rule functions
	for k, v := range RuleFunctions() {
		ruleFunctions[k] = v
	}

	return ruleFunctions
}

package main

// #cgo LDFLAGS: -L. -lz3
// #define MAX_ASSUMPTIONS 1024
// #include "../c/z3.h"
// #include <stdlib.h>
// void error_handler(Z3_context c, Z3_error_code e)
// {
// 		printf("Error code: %d\n", e);
// 	    fprintf(stderr,"BUG: %s.\n", "Incorrect use of Z3");
//      exit(1);
// }
// Z3_context mk_context_custom(Z3_config cfg, Z3_error_handler err)
// {
//     Z3_context ctx;
//     Z3_set_param_value(cfg, "model", "true");
//     ctx = Z3_mk_context(cfg);
//     Z3_set_error_handler(ctx, err);
//     return ctx;
// }
// Z3_context mk_context()
// {
//     Z3_config  cfg;
//     Z3_context ctx;
//     cfg = Z3_mk_config();
//     ctx = mk_context_custom(cfg, error_handler);
//     Z3_del_config(cfg);
//     return ctx;
// }
// Z3_solver mk_solver(Z3_context ctx)
// {
//   Z3_solver s = Z3_mk_solver(ctx);
//   Z3_solver_inc_ref(ctx, s);
//   return s;
// }
// void del_solver(Z3_context ctx, Z3_solver s)
// {
//   Z3_solver_dec_ref(ctx, s);
// }
// Z3_string check_sat(Z3_string p, Z3_string a)
// {
//     Z3_model m = 0;
//     Z3_context c = mk_context();
//     Z3_solver s = mk_solver(c);
//     Z3_ast_vector f = Z3_parse_smtlib2_string(c, p, 0,0,0,0,0,0);
//     for (unsigned i = 0; i < Z3_ast_vector_size(c, f); ++i) {
//         Z3_solver_assert(c, s, Z3_ast_vector_get(c, f, i));
//     }
//     Z3_ast_vector b = Z3_parse_smtlib2_string(c, a, 0,0,0,0,0,0);
//     Z3_ast assumptions[MAX_ASSUMPTIONS];
//     unsigned num_assumptions = Z3_ast_vector_size(c, b);
//     for (unsigned i = 0; i < num_assumptions; ++i) assumptions[i] = Z3_ast_vector_get(c, b, i);
//     Z3_lbool result = Z3_solver_check_assumptions(c, s, num_assumptions, assumptions);
//     switch (result) {
//     case Z3_L_FALSE:
//         printf("unsat\n");
//         Z3_ast_vector core = Z3_solver_get_unsat_core(c, s);
//         printf("vector size %d\n", Z3_ast_vector_size(c, core));
//          for (unsigned i = 0; i < Z3_ast_vector_size(c, core); ++i) {
//               printf("%s\n", Z3_ast_to_string(c, Z3_ast_vector_get(c, core, i)));
//          }
//         break;
//     case Z3_L_UNDEF:
//         printf("unknown\n");
//     m = Z3_solver_get_model(c, s);
//     if (m) Z3_model_inc_ref(c, m);
//         printf("potential model:\n%s\n", Z3_model_to_string(c, m));
//         break;
//     case Z3_L_TRUE:
//     m = Z3_solver_get_model(c, s);
//     if (m) Z3_model_inc_ref(c, m);
//         printf("sat\n%s\n", Z3_model_to_string(c, m));
//         break;
//     }
//     if (m) Z3_model_dec_ref(c, m);
//     del_solver(c, s);
//     //Z3_del_context(c);
// }
import "C"

import (
	"fmt"
	"go/ast"
	"go/parser"
)

// CheckSat check satisfiability
func CheckSat(p string, a string) string {
	return C.GoString(C.check_sat(C.CString(p), C.CString(a)))
}

// RuleVariablesSet maps recongnized rule variables to their SMT type TODO TODO TODO TODO TODO transfer this back to common/config.go
var RuleVariablesSet = map[string]string{ // Keep it in alphabetical order for easy reading
	"Action":        "String",
	"Amount":        "Real",
	"Balance":       "Real",
	"Destaccount":   "String",
	"Initiator":     "String",
	"Recipient":     "String",
	"Rule":          "String",
	"Ruletype":      "String", // this is not used currently, right?
	"Sourceaccount": "String",
}

// FuncSignatures maps a "rule" function to its SMT signature
var FuncSignatures = map[string]string{
	"f": "(Real Real) Real", // TODO bogus function just for testing
}

// map go operator to corresponding smtlib operator. Note: only needed for binary operators
func mapOp(op string) string {
	if op == "==" {
		return "="
	}

	return op
}

// AstWalk walks the abstract syntax tree of the expression (typically rule)
func AstWalk(n ast.Node) (string, []string, []string) {
	var s string
	vars := make([]string, 0)
	funcs := make([]string, 0)
	switch x := n.(type) {
	case *ast.BasicLit:
		s = x.Value
	case *ast.Ident:
		s = x.Name
		vars = append(vars, x.Name)
	case *ast.BinaryExpr:
		left, lvs, lfs := AstWalk(x.X)
		right, rvs, rfs := AstWalk(x.Y)
		op := mapOp(x.Op.String())
		s = "(" + op + " " + left + " " + right + ")"
		vars = append(append(vars, lvs...), rvs...)
		funcs = append(append(funcs, lfs...), rfs...)
	case *ast.UnaryExpr:
		ss, vs, fs := AstWalk(x.X)
		// add "0" to turn unary operator into binary
		s = "(" + x.Op.String() + " 0 " + ss + ")"
		vars = append(vars, vs...)
		funcs = append(funcs, fs...)
	case *ast.CallExpr:
		s = "(" + fmt.Sprint(x.Fun) + " "
		funcs = append(funcs, fmt.Sprint(x.Fun))
		args := x.Args
		for i, arg := range args {
			ss, vs, fs := AstWalk(arg)
			s += ss
			if i < len(args)-1 {
				s += " " // no "," between function arguments in z3
			}
			vars = append(vars, vs...)
			funcs = append(funcs, fs...)
		}
		s += ")"
	case *ast.ParenExpr:
		ss, vs, fs := AstWalk(x.X)
		s = ss
		vars = append(vars, vs...)
		funcs = append(funcs, fs...)
	default:
		fmt.Println("type ", x, " not supported")
	}

	return s, vars, funcs
}

// ToSMT uses output from ast walker to build SMT-LIB 2 code
func ToSMT(es []string, mvars, mfuncs map[string]bool) string {
	ret := ""
	// mvars should contain all the variables from all the es. store in map to avoid duplicates
	for v := range mvars {
		ret += "(declare-const " + v + " " + RuleVariablesSet[v] + ")\n"
	}

	// mfuncs should contain all the rule functions from all the es. store in map to avoid duplicates
	for f := range mfuncs {
		ret += "(declare-fun " + f + " " + FuncSignatures[f] + ")\n"
	}

	// now the es, the expressions (rules)
	for _, e := range es {
		ret += "(assert " + e + ")\n"
	}

	/////////////////////////////////////////////////////////////////////////ret += "(check-sat)\n(get-model)\n"

	return ret
}

// CreateSMTprogram converts a set of expressions (typically rules) into a SMT-LIB 2 program
func CreateSMTprogram(exprs []string) string {
	retExprs := make([]string, 0)
	mvars, mfuncs := make(map[string]bool), make(map[string]bool)
	for _, e := range exprs {
		astE, err := parser.ParseExpr(e)
		if err != nil {
			panic(err)
		}

		smtExpr, vars, funcs := AstWalk(astE)
		for _, v := range vars {
			mvars[v] = true
		}
		for _, f := range funcs {
			mfuncs[f] = true
		}
		retExprs = append(retExprs, smtExpr)
	}

	return ToSMT(retExprs, mvars, mfuncs)
}

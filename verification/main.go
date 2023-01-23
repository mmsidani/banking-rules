package main

import (
	"fmt"
)

// func findConflict(p string) string {
// 	///ctx := makeZ3Context()
// 	junk()
// 	//fmt.Println(ctx)
// 	return ""
// }

func main() {
	// src is the input for which we want to inspect the AST.
	neut := `0 == 0`
	src := `f(+5, ((Amount + 10) * 50)) < 10000`
	src2 := `Amount > 3000.0`
	//src3 := `Amount < 2000.0`
	src4 := `Amount < 10000`
	src5 := `Recipient=="ID12345"`
	src6 := `Balance < 100.0`
	//src7 := `Balance > 200.0`

	p := CreateSMTprogram([]string{neut})
	fmt.Println(p)

	//////////////////////c := MakeZ3Context()
	////////////////s := MakeZ3Solver(c)

	a := CreateSMTprogram([]string{src, src2, src4, src5, src6})
	fmt.Println(a)

	res := CheckSat(p, a)

	fmt.Println(res)
}

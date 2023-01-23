package main

// just an idea for now 9/6/18:
// 1) rule to require signerpubkey in signedpayload match one that initiator is known to have
// 2) compliance rule imposed by bank (exclude black listed entities, etc.)
// 3) so unlike all other rules, these are not account specific but bank specific

import (
	"os"
)

func main() {
	// get the arguments without the program name
	arg := os.Args[1:]

	if len(arg) != 1 {
		os.Exit(1)
	}

}

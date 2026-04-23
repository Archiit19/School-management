//go:build ignore

package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	pw := "DemoPass123!"
	if len(os.Args) > 1 {
		pw = os.Args[1]
	}
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(h))
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/stuphlabs/pullcord/authentication"
)

const maxUint16 = ^uint16(0)

func main() {
	var rawIterations uint
	flag.UintVar(
		&rawIterations,
		"iterations",
		uint(authentication.Pbkdf2MinIterations),
		"Iteration count for PBKDF2",
	)

	var password string
	flag.StringVar(
		&password,
		"password",
		"",
		"Password string, taken from stdin if empty or not specified",
	)

	flag.Parse()

	if rawIterations < uint(authentication.Pbkdf2MinIterations) {
		panic(fmt.Errorf(
			"Iteration count must be at least %d",
			authentication.Pbkdf2MinIterations,
		))
	} else if rawIterations > uint(maxUint16) {
		panic(fmt.Errorf(
			"Iteration count must be no greater than %d",
			maxUint16,
		))
	}
	iterations := uint16(rawIterations)

	if password == "" {
		_ = fmt.Fprintf(os.Stderr, "Password: ")

		_, err := fmt.Scanf("%s\n", &password)
		if err != nil {
			panic(err)
		}
	}

	hash, err := authentication.GetPbkdf2Hash(password, iterations)
	if err != nil {
		panic(err)
	}

	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(hash)
	if err != nil {
		panic(err)
	}
}

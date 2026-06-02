// Package main is the entrypoint for the gherkinator CLI. It delegates
// all command logic to the cmd sub-package.
package main

import "gherkinator/cmd/gherkinator/cmd"

func main() {
	cmd.Execute()
}

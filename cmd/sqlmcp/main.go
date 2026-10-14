package main

import (
	"fmt"
	"os"
)

var (
	version = "0.1.0-dev"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("SQLMCP version %s\n", version)
		os.Exit(0)
	}

	fmt.Fprintln(os.Stderr, "SQLMCP Server - Multi-dialect Database MCP Bridge")
	fmt.Fprintln(os.Stderr, "Status: Scaffolding concluído. Pronto para orquestração pelo /developer.")
}

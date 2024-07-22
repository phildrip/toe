package main

import (
	"flag"
	"fmt"
	"os"
	"toe/gen"
)

func main() {
	var outputFile string
	var disableFormatting bool

	flag.BoolVar(&disableFormatting, "no-fmt", false, "disable formatting of the output")
	flag.StringVar(&outputFile, "o", "", "output file name")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintf(
			os.Stderr,
			"Usage: %s [-no-fmt] -o <output.go> <input_directory> <interface>\n",
			os.Args[0])

		os.Exit(1)
	}

	inputDir := flag.Arg(0)
	interfaceName := flag.Arg(1)

	interfaceMethods, packageName, err := gen.FindInterface(inputDir, interfaceName)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding interface: %v\n", err)
		os.Exit(1)
	}

	if len(interfaceMethods) == 0 {
		fmt.Fprintf(os.Stderr, "Interface %s not found\n", interfaceName)
		os.Exit(1)
	}

	stubCode, err := gen.GenerateStubCode(interfaceName, interfaceMethods, packageName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating stub: %v\n", err)
		os.Exit(1)
	}

	if outputFile == "" {
		fmt.Println(stubCode)
	} else {
		err := os.WriteFile(outputFile, []byte(stubCode), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Stub generated in %s\n", outputFile)
	}
}

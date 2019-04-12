package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		compress   bool
		decompress bool
		useStdin   bool
	)
	flag.BoolVar(&compress, "c", false, "Compress")
	flag.BoolVar(&decompress, "d", false, "Decompress")

	flag.Parse()

	tail := flag.Args()
	if len(tail) > 0 {
		useStdin = true
	}

	if compress && decompress {
		fmt.Println("Cannot set both -c and -d")
		os.Exit(1)
	}

	if !compress && !decompress {
		// setting compress to true as default for the group
		compress = true
	}

	fmt.Println(compress)
	fmt.Println(decompress)
	fmt.Println(useStdin)
}

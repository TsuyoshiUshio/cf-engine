package main 

import (
	"fmt"
	"flag"
)

func main() {
	var fileName = flag.String("f", ".container-config.yml", "Configuration File: Path to Configuration file")

	flag.Parse()
	fmt.Println(*fileName)
} 

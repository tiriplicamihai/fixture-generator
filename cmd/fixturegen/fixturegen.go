package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/tiriplicamihai/fixture-generator/fixturegenerator"
)

var structName = flag.String("struct", "", "The struct you want to generate fixtures for")

func main() {
	flag.Parse()
	if *structName == "" {
		fmt.Println("You must specify a struct name.")
		return
	}

	rand.Seed(time.Now().Unix())

	structType, err := fixturegenerator.GetStruct(".", *structName)
	if err != nil {
		fmt.Println(err)
		return
	}
	imports, err := fixturegenerator.GetImports(".")
	if err != nil {
		fmt.Println(err)
		return
	}
	fixture := fixturegenerator.GetStructFixture(structType, imports, false, 1)

	fmt.Println(fixture)
}

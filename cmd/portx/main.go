package main

import (
	"log"
	"os"

	"github.com/alexperezortuno/portx/internal/bootstrap"
)

func main() {

	rt, err := bootstrap.New()

	if err != nil {
		log.Fatal(err)
	}

	if err := rt.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

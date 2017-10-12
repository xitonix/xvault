package obfuscate

import (
	"context"
	"log"
	"os"
	"time"
)

func ExampleEncoder_encode() {
	master, err := KeyFromPassword("password")

	if err != nil {
		log.Fatal(err)
	}

	input, err := os.Open("input.dat")
	if err != nil {
		log.Fatal(err)
	}

	output, err := os.Create("encoded.dat.xv")
	if err != nil {
		log.Fatal(err)
	}

	encoder := NewEncoder(1024, master, input, output)
	_, err = encoder.Encode()

	if err != nil {
		log.Fatal(err)
	}
}

func ExampleEncoder_encodeContext() {
	master, err := KeyFromPassword("password")

	if err != nil {
		log.Fatal(err)
	}

	input, err := os.Open("big_input.dat")
	if err != nil {
		log.Fatal(err)
	}

	output, err := os.Create("big_encoded.dat.xv")
	if err != nil {
		log.Fatal(err)
	}

	encoder := NewEncoder(1024, master, input, output)
	ctx, cancel := context.WithCancel(context.Background())
	//Queued the time consuming process of encoding a big file
	_, err = encoder.EncodeContext(ctx)

	go func(cancel context.CancelFunc) {
		<-time.After(5 * time.Second)
		cancel()
	}(cancel)

	if err != nil {
		log.Fatal(err)
	}
}

package obfuscate

import (
	"context"
	"log"
	"os"
	"time"
)

func ExampleDecoder_decode() {
	master, err := KeyFromPassword("password")

	if err != nil {
		log.Fatal(err)
	}

	input, err := os.Open("encoded.dat.xv")
	if err != nil {
		log.Fatal(err)
	}

	output, err := os.Create("decoded.dat")
	if err != nil {
		log.Fatal(err)
	}

	decoder := NewDecoder(1024, master, input, output)
	_, err = decoder.Decode()

	if err != nil {
		log.Fatal(err)
	}
}

func ExampleDecoder_decodeContext() {
	master, err := KeyFromPassword("password")

	if err != nil {
		log.Fatal(err)
	}

	input, err := os.Open("big_encoded.dat.xv")
	if err != nil {
		log.Fatal(err)
	}

	output, err := os.Create("big_decoded.dat")
	if err != nil {
		log.Fatal(err)
	}

	decoder := NewDecoder(1024, master, input, output)
	ctx, cancel := context.WithCancel(context.Background())

	//Queued the time consuming process of decoding a big file
	_, err = decoder.DecodeContext(ctx)

	go func(cancel context.CancelFunc) {
		<-time.After(5 * time.Second)
		cancel()
	}(cancel)

	if err != nil {
		log.Fatal(err)
	}
}

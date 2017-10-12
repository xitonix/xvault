package main

import (
	"fmt"
	"github.com/xitonix/xvault/obfuscate"
	"github.com/xitonix/xvault/taps"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {

	master, err := obfuscate.KeyFromPassword("password")
	if err != nil {
		log.Fatal(err)
	}

	pipe := obfuscate.NewPipe(10)
	if err != nil {
		log.Fatalln(err)
	}

	tap, err := taps.NewFilesystemTap("d:\\src", "d:\\target", 100*time.Millisecond, true, true, true)
	bucket, err := obfuscate.NewBucket(master, pipe, tap)

	if err != nil {
		log.Fatalln(err)
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range tap.Errors() {
			fmt.Println("Err: ", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for p := range tap.Progress() {
			if p.Status == obfuscate.Queued {
				fmt.Printf("Encrypting %s...\n", p.Input)
				continue
			}
			fmt.Printf("%s %s\n", p.Input, p.Status)
		}
	}()

	bucket.Open()

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	bucket.Close()
	fmt.Println("The bucket has been stopped successfully")
	wg.Wait()
}

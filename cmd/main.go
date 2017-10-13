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

	tap, err := taps.NewFilesystemTap("d:\\src", "d:\\target", 100*time.Millisecond, master, true, true)
	engine, err := obfuscate.NewEngine(10, true, tap)

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
		for p := range engine.Progress() {
			m := tap.ParseMetadata(p.Metadata)
			if p.Status == obfuscate.Queued {
				fmt.Printf("Encrypting %s...\n", m.Input.Name)
				continue
			}
			fmt.Printf("%s > %s %s\n", m.Input.Name, m.Output.Name, p.Status)
		}
	}()

	err = engine.Start()
	if err != nil {
		log.Fatal(err)
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	engine.Stop()
	fmt.Println("The engine has been stopped successfully")
	wg.Wait()
}

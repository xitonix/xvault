// Directory watcher tool
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/xitonix/xvault/obfuscate"
	"github.com/xitonix/xvault/taps"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {

	fmt.Print("Enter your password: ")
	password, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
	}

	master, err := obfuscate.KeyFromPassword(string(password))
	if err != nil {
		log.Fatal(err)
	}

	tap, err := taps.NewFileTap("src", "d:\\target", time.Second, master, true, true, true)

	if err != nil {
		log.Fatal(err)
	}

	engine := obfuscate.NewEngine(50, tap)
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
			fmt.Printf("%s > %s %s\n", p.Input.Name, p.Output.Name, p.Status)
		}
	}()

	engine.Start()

	go func() {
		signals := make(chan os.Signal)
		signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		<-signals
		engine.Stop()
	}()

	tap.Process()
	engine.Stop()
	wg.Wait()
	fmt.Println("The engine has been stopped successfully")
}

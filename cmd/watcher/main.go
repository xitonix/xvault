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

	fmt.Println("\nStarting the service...")

	tap, err := taps.NewDirectoryWatcherTap("src", "d:\\target", 100*time.Millisecond, master, true, false, true)

	if err != nil {
		log.Fatal(err)
	}

	engine := obfuscate.NewEngine(10, tap)
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

	fmt.Println("The service is up and running. Press Ctrl+C to stop it")

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	engine.Stop()
	fmt.Println("The engine has been stopped successfully")
	wg.Wait()
}

// Package xvault provides the functionality of encrypting/decrypting any Reader into one or more Writers.
// You can manually use the provided Encoder and Decoder types or automate the encryption/decryption tasks by implementing
// a new Tap and passing it to an engine object.
//
// Here is an example of using an engine object with a Directory Watcher Tap:
//
//	import (
//		"fmt"
//		"log"
//		"os"
//		"os/signal"
//		"sync"
//		"syscall"
//		"time"
//
//		"github.com/xitonix/xvault/obfuscate"
//		"github.com/xitonix/xvault/taps"
//	)
//
//	func main() {
//		master, err := obfuscate.KeyFromPassword("password")
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		tap, err := taps.NewDirectoryWatcherTap("src", "target", 100*time.Millisecond, master, true, true, true)
//
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		engine := obfuscate.NewEngine(10, tap)
//		wg := &sync.WaitGroup{}
//
//		wg.Add(1)
//		go func() {
//			defer wg.Done()
//				for err := range tap.Errors() {
//				fmt.Println("Error: ", err)
//			}
//		}()
//
//		wg.Add(1)
//		go func() {
//			defer wg.Done()
//			for p := range tap.Progress() {
//				fmt.Printf("%s > %s %s\n", p.Input.Name, p.Output.Name, p.Status)
//			}
//		}()
//
//		engine.Start()
//
//		signals := make(chan os.Signal)
//		signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
//		<-signals
//
//		engine.Stop()
//		wg.Wait()
//	}
package xvault

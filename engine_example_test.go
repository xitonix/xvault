package xvault

//import (
//	"fmt"
//	"log"
//	"os"
//	"os/signal"
//	"sync"
//	"syscall"
//	"time"
//
//	"github.com/xitonix/xvault/obfuscate"
//	"github.com/xitonix/xvault/taps"
//)

//func ExampleNone() {
//	master, err := obfuscate.KeyFromPassword("password")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	tap, err := taps.NewDirectoryWatcherTap("src", "target", 100*time.Millisecond, master, true, true, true)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	engine := obfuscate.NewEngine(10, tap)
//	wg := &sync.WaitGroup{}
//
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		for err := range tap.Errors() {
//			fmt.Println("Error: ", err)
//		}
//	}()
//
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		for p := range tap.Progress() {
//			fmt.Printf("%s > %s %s\n", p.Input.Name, p.Output.Name, p.Status)
//		}
//	}()
//
//	engine.Start()
//
//	signals := make(chan os.Signal)
//	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
//	<-signals
//	engine.Stop()
//	wg.Wait()
//}

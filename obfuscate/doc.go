// Package obfuscate implements the core functionality of the secure vault.
//
// The core component of obfuscate package is the Engine type which is responsible for
// processing any type of io.Reader and encrypt/decrypt it into one or more Writers.
//
// Every engine is connected to a pipe of work units from which it receives the requests.
// In order to flow the work units into the associated pipe, you need to implement a Tap and
// connect it to the engine by passing it to obfuscate.NewEngine(...) method.
//
//	tap, err := taps.YourImplementationOfTap(...)
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	engine := obfuscate.NewEngine(bufferSize, tap)
//
// Once you initialised the engine, you need to start it:
//	engine.Start()
//
//	signals := make(chan os.Signal)
//	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
//	<-signals
//
//	engine.Stop()
package obfuscate

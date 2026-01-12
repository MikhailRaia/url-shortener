package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	"github.com/MikhailRaia/url-shortener/internal/app"
	"github.com/MikhailRaia/url-shortener/internal/config"
)

var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func writeHeapProfile(path string) {
	f, err := os.Create(path)
	if err == nil {
		runtime.GC()
		pprof.WriteHeapProfile(f)
		_ = f.Close()
	}
}

func main() {
	if *memprofile != "" {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			writeHeapProfile(*memprofile)
			os.Exit(0)
		}()
	}

	cfg := config.NewConfig()

	if cfg.MaxProcs > 0 {
		runtime.GOMAXPROCS(cfg.MaxProcs)
	}

	application := app.NewApp(cfg)
	if err := application.Run(); err != nil {
		if *memprofile != "" {
			writeHeapProfile(*memprofile)
		}
		log.Fatalf("Error running application: %v", err)
	}

	if *memprofile != "" {
		writeHeapProfile(*memprofile)
	}
}

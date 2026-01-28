package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	"github.com/MikhailRaia/url-shortener/internal/app"
	"github.com/MikhailRaia/url-shortener/internal/config"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
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
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	if *memprofile != "" {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			writeHeapProfile(*memprofile)
			os.Exit(0)
		}()
	}

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

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

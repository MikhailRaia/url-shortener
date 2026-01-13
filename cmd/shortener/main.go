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
	buildVersion string
	buildDate    string
	buildCommit  string
)

var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

func orNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

func writeHeapProfile(path string) {
	f, err := os.Create(path)
	if err == nil {
		runtime.GC()
		pprof.WriteHeapProfile(f)
		_ = f.Close()
	}
}

func main() {
	fmt.Printf("Build version: %s\n", orNA(buildVersion))
	fmt.Printf("Build date: %s\n", orNA(buildDate))
	fmt.Printf("Build commit: %s\n", orNA(buildCommit))

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

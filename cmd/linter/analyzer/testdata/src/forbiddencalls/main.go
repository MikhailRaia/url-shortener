package forbiddencalls

import (
	"log"
	"os"
)

func main() {
	log.Fatal("allowed in main") // No want
	os.Exit(0)                   // No want
}

func init() {
	panic("panic forbidden even in init") // want "panic is forbidden"
	log.Fatal("forbidden in init")        // want "log.Fatal is forbidden outside main function"
	os.Exit(1)                            // want "os.Exit is forbidden outside main function"
}

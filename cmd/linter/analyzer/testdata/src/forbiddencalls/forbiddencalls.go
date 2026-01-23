package forbiddencalls

import (
	"log"
	"os"
)

// SomePanicFunction contains a panic call for analyzer testing.
func SomePanicFunction() {
	panic("this is forbidden") // want "panic is forbidden"
}

// SomeLogFatalFunction contains a log.Fatal call for analyzer testing.
func SomeLogFatalFunction() {
	log.Fatal("this is forbidden") // want "log.Fatal is forbidden outside main function"
}

// SomeOsExitFunction contains an os.Exit call for analyzer testing.
func SomeOsExitFunction() {
	os.Exit(1) // want "os.Exit is forbidden outside main function"
}

// AnotherPanicCall contains another panic call for analyzer testing.
func AnotherPanicCall() {
	panic("another panic") // want "panic is forbidden"
}

// MultipleCallsFunction contains multiple forbidden calls for analyzer testing.
func MultipleCallsFunction() {
	panic("panic 1")   // want "panic is forbidden"
	log.Fatal("fatal") // want "log.Fatal is forbidden outside main function"
	os.Exit(0)         // want "os.Exit is forbidden outside main function"
}

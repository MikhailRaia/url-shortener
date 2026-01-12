package forbiddencalls

import (
	"log"
	"os"
)

func SomePanicFunction() {
	panic("this is forbidden") // want "panic is forbidden"
}

func SomeLogFatalFunction() {
	log.Fatal("this is forbidden") // want "log.Fatal is forbidden outside main function"
}

func SomeOsExitFunction() {
	os.Exit(1) // want "os.Exit is forbidden outside main function"
}

func AnotherPanicCall() {
	panic("another panic") // want "panic is forbidden"
}

func MultipleCallsFunction() {
	panic("panic 1")   // want "panic is forbidden"
	log.Fatal("fatal") // want "log.Fatal is forbidden outside main function"
	os.Exit(0)         // want "os.Exit is forbidden outside main function"
}

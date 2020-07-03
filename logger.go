package easytls

import (
	"log"
	"os"
)

// NewDefaultLogger will initialize a new logger to the default used by this library.
// This will write to STDOUT, with no additional prefix beyond the date provided by
// log.Lstdflags.
func NewDefaultLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LstdFlags)
}

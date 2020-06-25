package easytls

import (
	"log"
	"os"
)

// NewDefaultLogger will initialize a new logger to the default used by this library.
func NewDefaultLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LstdFlags)
}

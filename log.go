package gosubmit

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "gosubmit ", log.LstdFlags)

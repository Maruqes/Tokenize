package Tokenize

import (
	"testing"
)

func TestThis(t *testing.T) {
	Initialize()
	InitListen("4242", "/success.html", "/cancel.html")
}

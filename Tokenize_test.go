package Tokenize

import (
	"Tokenize/database"
	"testing"
)

func TestThis(t *testing.T) {
	Init()
}

func TestPermission1(t *testing.T) {
	database.Init()
	err := CreatePermission("superadmin2", "all2:all")
	if err != nil {
		t.Error(err)
	}
}

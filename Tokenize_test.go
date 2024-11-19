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
	err := CreatePermission("balcao", "bar:write")
	if err != nil {
		t.Error(err)
	}
}

func TestPermission2(t *testing.T) {
	database.Init()
	err := DeletePermission(5)
	if err != nil {
		t.Error(err)
	}
}

func TestPermission3(t *testing.T) {
	database.Init()
	err := AddUserPermission(1, 1)
	if err != nil {
		t.Error(err)
	}
}

func TestPermission4(t *testing.T) {
	database.Init()
	err := RemoveUserPermission(1, 1)
	if err != nil {
		t.Error(err)
	}
}

func TestPermission5(t *testing.T) {
	database.Init()
	perms, err := GetUserPermissions(1)
	t.Log(perms)

	if err != nil {
		t.Error(err)
	}
	if len(perms) == 0 {
		t.Error("no permissions")
	}
	t.Log("Permissions len: ", len(perms))
	for _, perm := range perms {
		t.Log(perm)
	}
}

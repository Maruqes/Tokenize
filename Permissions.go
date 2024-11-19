package Tokenize

import (
	"Tokenize/database"
	"fmt"
)

func CreatePermission(name, permission string) error {
	permission_type, err := database.GetPermissionWithName(name)
	if permission_type.ID != -1 || err != nil {
		return fmt.Errorf("permission %s already exists", name)
	}

	permission_type, err = database.GetPermissionWithPermission(permission)
	if permission_type.ID != -1 || err != nil {
		return fmt.Errorf("permission %s already exists", permission)
	}
	fmt.Println("Creating permission")
	return nil
}

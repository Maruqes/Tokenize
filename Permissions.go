package Tokenize

import (
	"Tokenize/database"
	"fmt"
)

type permissions struct {
}

// only all:all perms can do this
func (*permissions) CreatePermission(name, permission string) error {
	permission_type, err := database.GetPermissionWithName(name)
	if permission_type.ID != -1 || err != nil {
		return fmt.Errorf("permission %s already exists", name)
	}

	permission_type, err = database.GetPermissionWithPermission(permission)
	if permission_type.ID != -1 || err != nil {
		return fmt.Errorf("permission %s already exists", permission)
	}
	fmt.Println("Creating permission")
	database.CreateNewPermission(name, permission)
	return nil
}

func (*permissions) DeletePermission(id int) error {
	exist := database.CheckPermissionID(id)
	if !exist {
		return fmt.Errorf("permission %d does not exist", id)
	}

	database.DeletePermissionWithID(id)
	return nil
}

func (*permissions) AddUserPermission(userID, permissionID int) error {
	exist_id := database.CheckIfUserIDExists(userID)
	if !exist_id {
		return fmt.Errorf("user %d does not exist", userID)
	}

	exist_perm := database.CheckPermissionID(permissionID)
	if !exist_perm {
		return fmt.Errorf("permission %d does not exist", permissionID)
	}

	exist := database.CheckUserPermission(userID, permissionID)
	if exist {
		return fmt.Errorf("user %d already has permission %d", userID, permissionID)
	}

	database.AddUserPermission(userID, permissionID)
	return nil
}

func (*permissions) RemoveUserPermission(userID, permissionID int) error {
	exist_id := database.CheckIfUserIDExists(userID)
	if !exist_id {
		return fmt.Errorf("user %d does not exist", userID)
	}

	exist_perm := database.CheckPermissionID(permissionID)
	if !exist_perm {
		return fmt.Errorf("permission %d does not exist", permissionID)
	}

	database.RemoveUserPermission(userID, permissionID)
	return nil
}

// only the own user or all:all perms can do this
func (*permissions) GetUserPermissions(userID int) ([]database.Permission, error) {
	exist_id := database.CheckIfUserIDExists(userID)
	if !exist_id {
		return []database.Permission{}, fmt.Errorf("user %d does not exist", userID)
	}

	return database.GetUserPermissions(userID)
}

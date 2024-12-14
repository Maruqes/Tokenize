package Permissions

import (
	"fmt"

	"github.com/Maruqes/Tokenize/database"
)

// only all:all perms can do this
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
	err = database.CreateNewPermission(name, permission)
	if err != nil {
		return fmt.Errorf("error creating permission %s", name)
	}
	return nil
}

func DeletePermission(id int) error {
	exist := database.CheckPermissionID(id)
	if !exist {
		return fmt.Errorf("permission %d does not exist", id)
	}

	err := database.DeletePermissionWithID(id)
	if err != nil {
		return fmt.Errorf("error deleting permission %d", id)
	}
	return nil
}

func GetPermissions() ([]database.Permission, error) {
	return database.GetPermissions()
}

func AddUserPermission(userID, permissionID int) error {
	exist_id, err := database.CheckIfUserIDExists(userID)
	if err != nil || !exist_id {
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

func RemoveUserPermission(userID, permissionID int) error {
	exist_id, err := database.CheckIfUserIDExists(userID)
	if err != nil || !exist_id {
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
func GetUserPermissions(userID int) ([]database.Permission, error) {
	exist_id, err := database.CheckIfUserIDExists(userID)
	if err != nil || !exist_id {
		return []database.Permission{}, fmt.Errorf("user %d does not exist", userID)
	}

	return database.GetUserPermissions(userID)
}

func HasPermission(userID int, requiredPermission string) bool {
	userPermissions, err := database.GetUserPermissions(userID)
	if err != nil {
		return false
	}

	for _, perm := range userPermissions {
		if perm.Permission == requiredPermission {
			return true
		}
	}
	return false
}

func GetAllUsersPermissions() ([]database.Permission, error) {
	return database.GetAllUsersPermissions()
}

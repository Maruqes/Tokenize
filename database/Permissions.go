package database

import "log"

type Permission struct {
	ID         int
	Name       string
	Permission string
}

func CreatePermissionsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS permissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		permission TEXT NOT NULL,
		name TEXT NOT NULL,
	);`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

	query = `
	CREATE TABLE IF NOT EXISTS user_permissions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		permission_id TEXT NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id),
		FOREIGN KEY(permission_id) REFERENCES permissions(id)
	);`

	_, err = db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func CreateNewPermission(name, permission string) error {
	query := `INSERT INTO permissions (name, permission) VALUES (?, ?);`
	_, err := db.Exec(query, name, permission)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func EditPermissionWithID(id int, name, permission string) error {
	query := `UPDATE permissions SET name = ?, permission = ? WHERE id = ?;`
	_, err := db.Exec(query, name, permission, id)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func DeletePermissionWithID(id int) error {
	query := `DELETE FROM permissions WHERE id = ?;`
	_, err := db.Exec(query, id)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func AddUserPermission(userID int, permission Permission) error {
	query := `INSERT INTO user_permissions (user_id, permission_id) VALUES (?, ?);`
	_, err := db.Exec(query, userID, permission.ID)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func RemoveUserPermission(userID int, permission Permission) error {
	query := `DELETE FROM user_permissions WHERE user_id = ? AND permission_id = ?;`
	_, err := db.Exec(query, userID, permission.ID)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func GetUserPermissions(userID int) ([]Permission, error) {
	query := `
	SELECT permissions.id, permissions.name, permissions.permission
	FROM permissions
	JOIN user_permissions ON permissions.id = user_permissions.permission_id
	WHERE user_permissions.user_id = ?;
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var permissions []Permission
	for rows.Next() {
		var permission Permission
		err = rows.Scan(&permission.ID, &permission.Name, &permission.Permission)
		if err != nil {
			log.Fatal(err)
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

func GetPermissionWithID(id int) (Permission, error) {
	query := `SELECT id, name, permission FROM permissions WHERE id = ?;`
	row := db.QueryRow(query, id)
	var permission Permission
	err := row.Scan(&permission.ID, &permission.Name, &permission.Permission)
	if err != nil {
		log.Fatal(err)
	}
	return permission, nil
}

func CheckUserPermission(userID int, permissionID int) bool {
	query := `SELECT id FROM user_permissions WHERE user_id = ? AND permission_id = ?;`
	row := db.QueryRow(query, userID, permissionID)
	var result int
	err := row.Scan(&result)
	return err == nil
}

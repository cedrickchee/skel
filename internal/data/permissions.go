package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// Define a Permissions slice, which we will use to will hold the permission
// codes (like "movies:read" and "movies:write") for a single user.
type Permissions []string

// Add a helper method to check whether the Permissions slice contains a specific
// permission code.
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

// PermissionModel struct type which wraps a sql.DB connection pool.
type PermissionModel struct {
	DB *sql.DB
}

// GetAllForUser method returns all permission codes for a specific user in a
// Permissions slice.
func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
		SELECT permissions.code
		FROM permissions
		INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
		INNER JOIN users ON users_permissions.user_id = users.id
		WHERE users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions

	for rows.Next() {
		var permission string

		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// AddForUser adds the provided permission codes for a specific user. Notice
// that we're using a variadic parameter for the codes so that we can assign
// multiple permissions in a single call.
func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	query := `
        INSERT INTO users_permissions
        SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, pq.Array(codes))
	return err
}

// Mock models

type userPermissions struct {
	userID      int64
	permissions Permissions
}

var mockUserPermissions = []userPermissions{
	{userID: mockUser.ID, permissions: []string{"movies:read", "movies:write"}},
	{userID: 2, permissions: []string{"movies:read"}},
}

type MockPermissionModel struct{}

// GetAllForUser returns all mock permission codes for a specific user.
func (m MockPermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	var permissions Permissions

	for i := range mockUserPermissions {
		if mockUserPermissions[i].userID == userID {
			permissions = mockUserPermissions[i].permissions
			return permissions, nil
		}
	}

	return nil, ErrRecordNotFound
}

// AddForUser adds the provided permission codes for a specific user.
func (m MockPermissionModel) AddForUser(userID int64, codes ...string) error {
	return nil
}

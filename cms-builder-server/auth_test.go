package builder_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	builder "github.com/frangdelsolar/cms-builder/cms-builder-server"
	th "github.com/frangdelsolar/cms-builder/cms-builder-server/test_helpers"
	"github.com/stretchr/testify/assert"
)

// TestRegisterUserController tests the RegisterUser endpoint by creating a new  user and
// verifying the response to make sure the user was created correctly.
func TestRegisterUserController(t *testing.T) {
	t.Log("Testing VerifyUser")
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err, "GetDefaultEngine should not return an error")

	newUserData := builder.RegisterUserInput{
		Name:     th.RandomName(),
		Email:    th.RandomEmail(),
		Password: th.RandomPassword(),
	}

	bodyBytes, err := json.Marshal(newUserData)
	assert.NoError(t, err)

	header := http.Header{
		"Content-Type": []string{"application/json"},
	}

	responseWriter := th.MockWriter{}
	registerUserRequest := &http.Request{
		Method: http.MethodPost,
		Header: header,
		Body:   io.NopCloser(bytes.NewBuffer(bodyBytes)),
	}

	t.Log("Registering user")
	e.Engine.RegisterVisitorController(&responseWriter, registerUserRequest)

	t.Log("Testing Response")
	createdUser := builder.User{}
	response, err := builder.ParseResponse(responseWriter.Buffer.Bytes(), &createdUser)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, createdUser.Name, newUserData.Name)

	t.Log("Testing Verification token")
	accessToken, err := th.LoginUser(&newUserData)
	assert.NoError(t, err)

	t.Log("Verifying user")
	retrievedUser, err := e.Engine.VerifyUser(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, createdUser.GetIDString(), retrievedUser.GetIDString())

	t.Log("Rolling back user registration")
	e.Firebase.RollbackUserRegistration(context.Background(), createdUser.FirebaseId)
}

// TestAppendRoleToUser tests the AppendRoleToUser method for various scenarios.
// It verifies that roles are correctly appended to a user's roles field in the database.
// Test cases include situations where the user is not found, the user has no roles,
// the user already has the role, and the user does not have the role.
// It checks for expected errors and verifies that the roles are updated as expected.
func TestAppendRoleToUser(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err)

	adminRole := builder.Role("admin")
	moderatorRole := builder.Role("moderator")

	testCases := []struct {
		name          string
		userRoles     []builder.Role
		newRole       builder.Role
		expectedRoles string
		expectedErr   bool
	}{
		{
			name:          "User not found in database",
			userRoles:     []builder.Role{},
			newRole:       adminRole,
			expectedRoles: "admin",
			expectedErr:   true,
		},
		{
			name:          "User has no roles",
			userRoles:     []builder.Role{},
			newRole:       adminRole,
			expectedRoles: "admin",
			expectedErr:   false,
		},
		{
			name:          "User already has the role",
			userRoles:     []builder.Role{adminRole},
			newRole:       adminRole,
			expectedRoles: "admin",
			expectedErr:   true,
		},
		{
			name:          "User does not have the role",
			userRoles:     []builder.Role{adminRole},
			newRole:       moderatorRole,
			expectedRoles: "admin,moderator",
			expectedErr:   false,
		},
	}

	var systemUser = &builder.User{
		ID:    uint(9923142453952396459),
		Name:  "System",
		Email: "system@system",
	}

	for ix, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			roles := ""
			for _, role := range tc.userRoles {
				roles += string(role)
			}
			user := builder.User{
				Name:  th.RandomName(),
				Email: th.RandomEmail(),
				Roles: roles,
			}
			e.DB.Create(&user, systemUser)

			// first test should pass an invalid id.
			if ix == 0 {
				user.ID = uint(999999991231231234)
			}

			err := e.Engine.AppendRoleToUser(user.GetIDString(), tc.newRole)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				var savedUser builder.User
				e.DB.DB.First(&savedUser, user.ID)
				t.Log("Saved user", savedUser)
				assert.Equal(t, tc.expectedRoles, savedUser.Roles)
			}
		})
	}
}

// TestRemoveRoleFromUser tests the RemoveRoleFromUser method for various scenarios.
// It verifies that roles are correctly removed from a user's roles field in the database.
// Test cases include situations where the user is not found, the user has no roles,
// the user does not have the role, and the user has the role.
// It checks for expected errors and verifies that the roles are updated as expected.
func TestRemoveRoleFromUser(t *testing.T) {
	e, err := th.GetDefaultEngine()
	assert.NoError(t, err)

	adminRole := builder.Role("admin")
	moderatorRole := builder.Role("moderator")

	testCases := []struct {
		name          string
		userRoles     []builder.Role
		roleToRemove  builder.Role
		expectedRoles string
		expectedErr   bool
	}{
		{
			name:          "User not found in database",
			userRoles:     []builder.Role{},
			roleToRemove:  adminRole,
			expectedRoles: "",
			expectedErr:   true,
		},
		{
			name:          "User has no roles",
			userRoles:     []builder.Role{},
			roleToRemove:  adminRole,
			expectedRoles: "",
			expectedErr:   false,
		},
		{
			name:          "User does not have the role",
			userRoles:     []builder.Role{adminRole},
			roleToRemove:  moderatorRole,
			expectedRoles: "admin",
			expectedErr:   false,
		},
		{
			name:          "User has the role",
			userRoles:     []builder.Role{adminRole},
			roleToRemove:  adminRole,
			expectedRoles: "",
			expectedErr:   false,
		},

		{
			name:          "User has multiple roles",
			userRoles:     []builder.Role{adminRole, moderatorRole},
			roleToRemove:  adminRole,
			expectedRoles: "moderator",
			expectedErr:   false,
		},
	}

	var systemUser = &builder.User{
		ID:    uint(9923142453952396459),
		Name:  "System",
		Email: "system@system",
	}

	for ix, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			roles := ""
			for _, role := range tc.userRoles {
				if roles != "" {
					roles += ","
				}
				roles += string(role)
			}
			user := builder.User{
				Name:  th.RandomName(),
				Email: th.RandomEmail(),
				Roles: roles,
			}
			e.DB.Create(&user, systemUser)

			// first test should pass an invalid id.
			if ix == 0 {
				user.ID = uint(999999991231231234)
			}

			err := e.Engine.RemoveRoleFromUser(user.GetIDString(), tc.roleToRemove)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				var savedUser builder.User
				e.DB.DB.First(&savedUser, user.ID)
				t.Log("Saved user", savedUser)
				assert.Equal(t, tc.expectedRoles, savedUser.Roles)
			}
		})
	}
}

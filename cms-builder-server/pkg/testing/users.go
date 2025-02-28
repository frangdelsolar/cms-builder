package testing

import "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"

func CreateAllAllowedUser() *models.User {
	return CreateUser("All Allowed User", AllAllowedRole.S())
}

func CreateAdminUser() *models.User {
	return CreateUser("Admin User", models.AdminRole.S())
}

func CreateGodUser() *models.User {
	return CreateUser("God User", models.AdminRole.S())
}

func CreateVisitorUser() *models.User {
	return CreateUser("Visitor User", models.VisitorRole.S())
}

func CreateNoRoleUser() *models.User {
	return CreateUser("No Role User", "")
}

func CreateSchedulerUser() *models.User {
	return CreateUser("Scheduler User", models.SchedulerRole.S())
}

func CreateUser(name string, roles string) *models.User {

	name += " - " + RandomString(4)

	return &models.User{
		ID:    RandomUint(),
		Name:  name,
		Email: RandomEmail(),
		Roles: roles,
	}

}

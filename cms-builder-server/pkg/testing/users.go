package testing

import "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/models"

func CreateAllAllowedUser() *authModels.User {
	return CreateUser("All Allowed User", AllAllowedRole.S())
}

func CreateSystemUser() *authModels.User {
	return CreateUser("System User", "")
}

func CreateAdminUser() *authModels.User {
	return CreateUser("Admin User", models.AdminRole.S())
}

func CreateGodUser() *authModels.User {
	return CreateUser("God User", models.AdminRole.S())
}

func CreateVisitorUser() *authModels.User {
	return CreateUser("Visitor User", models.VisitorRole.S())
}

func CreateNoRoleUser() *authModels.User {
	return CreateUser("No Role User", "")
}

func CreateSchedulerUser() *authModels.User {
	return CreateUser("Scheduler User", models.SchedulerRole.S())
}

func CreateUser(name string, roles string) *authModels.User {

	name += " - " + RandomString(4)

	return &authModels.User{
		ID:    RandomUint(), // assing a random ID if we don't actually write the user to the database
		Name:  name,
		Email: RandomEmail(),
		Roles: roles,
	}

}

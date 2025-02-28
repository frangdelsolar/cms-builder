package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	orc "github.com/frangdelsolar/cms-builder/cms-builder-server"
	"github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/postman"
	mgr "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/resource-manager"
)

func main() {

	environment := flag.String("env", "test", "environemnt")
	runPostman := flag.Bool("postman", false, "Generate Postman files")
	flag.Parse()

	loadEnv(*environment)

	if !*runPostman {
		fmt.Print("Nothing to be done here...")
		os.Exit(0)
	}

	o, err := orc.NewOrchestrator()
	if err != nil {
		panic(err)
	}

	appName := o.Config.GetString(orc.EnvKeys.AppName)
	baseUrl := o.Config.GetString(orc.EnvKeys.BaseUrl)
	adminEmail := o.Config.GetString(orc.EnvKeys.AdminEmail)
	adminPassword := o.Config.GetString(orc.EnvKeys.AdminPassword)
	firebaseApiKey := o.Config.GetString(orc.EnvKeys.FirebaseApiKey)
	resources := []mgr.Resource{}

	for _, r := range o.ResourceManager.Resources {
		resources = append(resources, *r)
	}

	err = postman.ExportPostman(
		appName,
		*environment,
		baseUrl,
		adminEmail,
		adminPassword,
		firebaseApiKey,
		resources,
	)
	if err != nil {
		panic(err)
	}

}

func loadEnv(env string) {
	file := "." + env + ".env"

	err := godotenv.Load(file)
	if err != nil {
		fmt.Printf("Error loading .env file. %v\n", err)
		panic(err)
	}
}

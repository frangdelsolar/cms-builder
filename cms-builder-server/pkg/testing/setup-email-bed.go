package testing

import (
	"os"

	"github.com/joho/godotenv"

	emailPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/email"
	loggerTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/logger/types"
)

type EmailTestBedUtils struct {
	Logger      *loggerTypes.Logger
	EmailSender *emailPkg.EmailSender
}

func SetupEmailTestBed() *EmailTestBedUtils {

	godotenv.Load(".test.env")

	log := NewTestLogger()

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	smtpSender := os.Getenv("SMTP_SENDER")
	smtpFromName := os.Getenv("SMTP_FROM_NAME")

	sender := emailPkg.NewEmailSender(smtpHost, smtpPort, smtpUser, smtpPass, smtpSender, smtpFromName)

	return &EmailTestBedUtils{
		Logger:      log,
		EmailSender: sender,
	}
}

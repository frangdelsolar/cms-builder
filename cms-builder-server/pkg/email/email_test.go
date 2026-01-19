package email_test

import (
	"testing"

	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	"github.com/stretchr/testify/assert"
)

func TestEmailSenderSendEmail(t *testing.T) {
	bed := testPkg.SetupEmailTestBed()

	bed.Logger.Info().Interface("EmailSender", bed.EmailSender).Msg("EmailSender configured")

	recipientEmail := "frangdelsolar@gmail.com"
	subject := "Test Email - Desarrollo Psicositio"
	body := "<h1>Correo de Prueba</h1><p>Verificando remitentes y overrides.</p>"

	t.Run("Send with Default FromName", func(t *testing.T) {
		// Usa el FromName configurado en bed.EmailSender (desde .test.env)
		err := bed.EmailSender.SendEmail([]string{recipientEmail}, subject, body, "")

		assert.NoError(t, err)
	})

	t.Run("Send with Override FromName", func(t *testing.T) {
		// Forzamos un nombre distinto para este envío específico
		overrideName := "Soporte Psicositio"
		err := bed.EmailSender.SendEmail([]string{recipientEmail}, subject, body, overrideName)

		assert.NoError(t, err)
	})

	t.Run("Empty Recipient List Returns Error", func(t *testing.T) {
		err := bed.EmailSender.SendEmail([]string{}, subject, body, "")
		assert.Error(t, err)
	})
}

package email_test

import (
	"testing"

	testPkg "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/testing"
	"github.com/stretchr/testify/assert"
)

// TestEmailSenderSendEmail verifies email sending functionality using SMTP configuration from .test.env.
// This test requires valid SMTP credentials and server availability to pass.
func TestEmailSenderSendEmail(t *testing.T) {
	bed := testPkg.SetupEmailTestBed()

	bed.Logger.Info().Interface("EmailSender", bed.EmailSender).Msg("EmailSender configured")

	recipientEmail := "frangdelsolar@gmail.com"
	subject := "Test Email - Desarrollo Psicositio (Go Backend)"
	body := "<h1>Correo de Prueba Exitoso</h1><p>Este mensaje confirma que el servicio de correo de <b>Golang</b> está funcionando correctamente.</p>"

	t.Run("Successful Email Send with HTML Content", func(t *testing.T) {
		err := bed.EmailSender.SendEmail([]string{recipientEmail}, subject, body)

		assert.NoError(t, err, "Email sending failed. Verify SMTP credentials in .test.env and recipient address.")

		if err != nil {
			bed.Logger.Error().Err(err).Msg("Failed to send email. Check SMTP configuration in .test.env.")
		}
	})

	t.Run("Empty Recipient List Returns Error", func(t *testing.T) {
		err := bed.EmailSender.SendEmail([]string{}, subject, body)
		assert.Error(t, err, "Expected error when attempting to send without recipients")
	})
}

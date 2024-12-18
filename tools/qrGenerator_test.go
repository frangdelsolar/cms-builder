package tools_test

import (
	"testing"

	"github.com/frangdelsolar/cms/tools"
	"github.com/stretchr/testify/assert"
)

func TestGenerateQR(t *testing.T) {
	tools.GenerateQR()
	assert.FileExists(t, "qr.png")
	// os.Remove("repo-qrcode.jpeg")
}

package types

import "github.com/rs/zerolog"

// Logger wraps a zerolog.Logger instance with additional convenience methods
type Logger struct {
	*zerolog.Logger
}

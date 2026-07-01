package settings

import (
	apperrors "github.com/isyll/go-api-starter/internal/errors"
	"github.com/isyll/go-api-starter/internal/errors/codes"
)

// ErrSettingsNotFound - 404. No settings row for the user.
var ErrSettingsNotFound = apperrors.NotFound(codes.SettingsNotFound, "settings.not_found")

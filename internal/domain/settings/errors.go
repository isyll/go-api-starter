package settings

import (
	apperrors "github.com/isyll/go-grpc-starter/internal/errors"
	"github.com/isyll/go-grpc-starter/internal/errors/codes"
)

var ErrSettingsNotFound = apperrors.NotFound(codes.SettingsNotFound, "settings.not_found")

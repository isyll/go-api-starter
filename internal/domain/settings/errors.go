package settings

import (
	"github.com/isyll/go-grpc-starter/internal/errs"
	"github.com/isyll/go-grpc-starter/internal/errs/codes"
)

var ErrSettingsNotFound = errs.NotFound(codes.SettingsNotFound, "settings.not_found")

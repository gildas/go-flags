package flags

import (
	"net/http"

	"github.com/gildas/go-errors"
)

var (
	InvalidEnumValue = errors.NewSentinel(http.StatusBadRequest, "error.value.invalid", "Flag value \"%s\" in invalid. Expected values are %s")
)

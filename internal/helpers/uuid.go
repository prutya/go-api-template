package helpers

import (
	"errors"

	"github.com/gofrs/uuid/v5"
)

var ErrInvalidV7UUID = errors.New("invalid v7 uuid")

func ValidateUUIDV7(uuidString string) error {
	uuid, err := uuid.FromString(uuidString)
	if err != nil {
		return ErrInvalidV7UUID
	}

	if uuid.Version() != 7 {
		return ErrInvalidV7UUID
	}

	return nil
}

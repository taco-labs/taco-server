package payple

import (
	"errors"
)

type paypleExtensionOption func(*paypleExtension)

func WithDriverSEttlementApp(driverApp driversettlementApp) paypleExtensionOption {
	return func(pe *paypleExtension) {
		pe.app.driversettlement = driverApp
	}
}

func (p paypleExtension) validate() error {
	if p.app.driversettlement == nil {
		return errors.New("payple driver extension need driver app")
	}

	return nil
}

func NewPaypleExtension(opts ...paypleExtensionOption) (*paypleExtension, error) {
	extension := &paypleExtension{}

	for _, opt := range opts {
		opt(extension)
	}

	return extension, extension.validate()
}

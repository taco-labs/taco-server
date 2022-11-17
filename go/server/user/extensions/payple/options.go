package payple

import "errors"

type paypleExtensionOption func(*paypleExtension)

func WithDomain(domain string) paypleExtensionOption {
	return func(pe *paypleExtension) {
		pe.domain = domain
	}
}

func WithPayplePaymentApp(app userApp) paypleExtensionOption {
	return func(pe *paypleExtension) {
		pe.app.userApp = app
	}
}

func (p paypleExtension) validate() error {
	if p.renderer == nil {
		return errors.New("payple extension need renderer")
	}

	if p.app.userApp == nil {
		return errors.New("payple extension need payple payment app")
	}

	if p.domain == "" {
		return errors.New("payple extension need domain")
	}

	return nil
}

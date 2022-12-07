package payple

import "errors"

type paypleExtensionOption func(*paypleExtension)

func WithDomain(domain string) paypleExtensionOption {
	return func(pe *paypleExtension) {
		pe.domain = domain
	}
}

func WithPayplePaymentApp(app paymentApp) paypleExtensionOption {
	return func(pe *paypleExtension) {
		pe.app.paymentApp = app
	}
}

func WithEnv(env string) paypleExtensionOption {
	return func(pe *paypleExtension) {
		pe.env = env
	}
}

func (p paypleExtension) validate() error {
	if p.renderer == nil {
		return errors.New("payple extension need renderer")
	}

	if p.app.paymentApp == nil {
		return errors.New("payple extension need payple payment app")
	}

	if p.domain == "" {
		return errors.New("payple extension need domain")
	}

	if p.env == "" {
		return errors.New("payple extension need execution environment")
	}

	return nil
}

package payment

import (
	"context"
	"fmt"
)

func (p paymentApp) Start(ctx context.Context) error {
	go p.loop(ctx)
	return nil
}

func (p paymentApp) Stop(ctx context.Context) error {
	<-p.waitCh
	return nil
}

func (p paymentApp) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("shutting down [Payment App Consumer] stream...")
			p.waitCh <- struct{}{}
			return
		default:
		}
	}
}

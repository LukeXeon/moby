//go:build !linux || android

package libnetwork

func (c *Controller) selectFirewallBackend() error {
	return nil
}

func (c *Controller) setupPlatformFirewall() {}

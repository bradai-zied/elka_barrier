package elka

func (c *Controller) ChangeState() error {
	return c.SendCommand(0x00, 0x00) // Reverse Barrier Position
}

func (c *Controller) Open() error {
	return c.SendCommand(0x01, 0x00) // BT command, Pulse function
}

func (c *Controller) Close() error {
	return c.SendCommand(0x02, 0x00) // BZ command, Pulse function
}

func (c *Controller) LockOpen() error {

	return c.SendCommand(0x01, 0x01) // BA command, Activate function
}

func (c *Controller) LockClosed() error {
	return c.SendCommand(0x02, 0x01) // BZ command, Activate function
}

func (c *Controller) Unlock() error {
	err := c.SendCommand(0x01, 0x02) // BA command, Deactivate function
	if err != nil {
		return err
	}
	return c.SendCommand(0x02, 0x02) // BZ command, Deactivate function
}

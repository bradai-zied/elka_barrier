package elka

import "fmt"

func (c *Controller) GetBarrierStatus() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.BarrierStatus[:], nil
}

func (c *Controller) GetBarrierIP() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Barrierip
}

func (c *Controller) GetServiceCounter() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return []byte{
		byte(c.serviceCounter),
		byte(c.serviceCounter >> 8),
		byte(c.serviceCounter >> 16),
		byte(c.serviceCounter >> 24),
	}, nil
}

func (c *Controller) GetMaintenanceCounter() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return []byte{
		byte(c.maintenanceCounter),
		byte(c.maintenanceCounter >> 8),
		byte(c.maintenanceCounter >> 16),
		byte(c.maintenanceCounter >> 24),
	}, nil
}

func (c *Controller) GetGateState() (byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.gateState, nil
}

func (c *Controller) GetErrorMemory(index int) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if index < 0 || index >= len(c.errorMemory) {
		return nil, fmt.Errorf("invalid error memory index: %d", index)
	}
	return c.errorMemory[index][:], nil
}

func (c *Controller) GetVehicleCounter() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return []byte{
		byte(c.vehicleCounter),
		byte(c.vehicleCounter >> 8),
		byte(c.vehicleCounter >> 16),
		byte(c.vehicleCounter >> 24),
	}, nil
}

func (c *Controller) GetBarrierPosition() (byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return byte(c.BarrierPosition), nil
}

func (c *Controller) GetMotorStatus() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.motorStatus, nil
}

func (c *Controller) GetDebugLoops() ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.debugLoops, nil
}

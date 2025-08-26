package queue

// type PaymentQueue struct {
// 	q chan
// }

// func NewInstanceHealth() *InstanceHealth {
// 	return &InstanceHealth{
// 		Health: &types.ProcessorHealth{},
// 	}
// }

// func (h *InstanceHealth) GetHealth() *types.ProcessorHealth {
// 	h.mu.Lock()
// 	defer h.mu.Unlock()
// 	return h.Health
// }

// func (h *InstanceHealth) SetHealth(failing bool, minResp int) {
// 	h.mu.Lock()
// 	defer h.mu.Unlock()
// 	h.Health.Failing = failing
// 	h.Health.MinResponseTime = minResp
// 	h.Health.LastChecked = time.Now().UTC()
// }

// func (h *InstanceHealth) Expired(duration time.Duration) bool {
// 	h.mu.Lock()
// 	defer h.mu.Unlock()
// 	return time.Since(h.Health.LastChecked) > time.Duration(duration)
// }

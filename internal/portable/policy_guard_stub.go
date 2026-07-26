//go:build !windows

package portable

type policyGuard struct{}

func newPolicyGuard() *policyGuard { return &policyGuard{} }

func (g *policyGuard) Isolate() {}

func (g *policyGuard) Restore() {}

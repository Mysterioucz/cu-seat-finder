// Package identity holds the IdentityProvider boundary from ADR-004: student/staff
// login goes through a mock OIDC provider today and CU SSO later, without the
// User service caring which one is behind the interface.
package identity

import "errors"

type IdentityProvider interface {
	// Authenticate verifies a student/staff identity and returns their display name.
	// The mock trusts the client-supplied ID; a real OIDC provider would exchange
	// an auth code/token instead.
	Authenticate(studentID, nameHint string) (name string, err error)
}

// MockProvider is the dev stand-in for CU SSO (ADR-004). Swap for a real OIDC
// client once CU SSO credentials are issued; the User service code doesn't change.
type MockProvider struct{}

func (MockProvider) Authenticate(studentID, nameHint string) (string, error) {
	if studentID == "" {
		return "", errors.New("student_id required")
	}
	if nameHint != "" {
		return nameHint, nil
	}
	return "Student " + studentID, nil
}

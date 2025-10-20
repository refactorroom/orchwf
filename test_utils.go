package orchwf

// Helper function to compare maps
func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// Test error type
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

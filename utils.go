package orchwf

// Helper utility functions

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

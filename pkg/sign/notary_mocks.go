package sign

// MockParseReference is a mock function for ParseReferenceFunc
func MockParseReference(image string) (Reference, error) {
	return image, nil // In a simple case, we return the string itself as Reference
}

// MockGetImage is a mock function for GetImageFunc
func MockGetImage(_ Reference) (Image, error) {
	// We return a mocked Image object with predefined values
	return &SimpleImage{
		ManifestData: Manifest{
			Config: struct {
				Digest struct {
					Hex string
				}
				Size int64
			}{
				Digest: struct {
					Hex string
				}{
					Hex: "abc123def456",
				},
				Size: 12345678,
			},
		},
	}, nil
}

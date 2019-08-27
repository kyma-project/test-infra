package tester

// DO NOT EDIT. THIS FILE IS GENERATED

// List of currently supported releases
var (
	Release14 = mustParse("1.4")
	Release13 = mustParse("1.3")
	Release12 = mustParse("1.2")
)

// GetAllKymaReleaseBranches returns all supported kyma release branches
func GetAllKymaReleases() []*SupportedRelease {
	return []*SupportedRelease{
		Release14,
		Release13,
		Release12,
	}
}



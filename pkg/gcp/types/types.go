package types

// GCPBucketMetadata hnews metadata about Google Cloud bucket instance.
// Fields names are meaningfully so are easy to use in composition types.
type GCPBucketMetadata struct {
	GCPBucketName      *string `json:"gcpBucketName,omitempty"`
	GCPBucketDirectory *string `json:"gcpBucketDirectory,omitempty"`
}

// GCPProjectMetadata hnews metadata about Google Cloud project.
// Fields names are meaningfully so are easy to use in composition types.
type GCPProjectMetadata struct {
	GCPProject *string `json:"gcpProject,omitempty"`
}

package firestore

import (
	"cloud.google.com/go/firestore"
)

// Client wraps google firestore client and provide additional methods.
type Client struct {
	*firestore.Client
}
# (2025-03-04)
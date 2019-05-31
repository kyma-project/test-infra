package gcscleaner

func (r Cleaner) ExtractTimestampSuffix(name string) *string {
	return r.extractTimestampSuffix(name)
}

func (r Cleaner) ShouldDeleteBucket(name string, now int64) bool {
	return r.shouldDeleteBucket(name, now)
}

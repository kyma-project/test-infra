package gcscleaner

import "regexp"

func ExtractTimestampSuffix(name string, regTimestampSuffix regexp.Regexp) *string {
	return extractTimestampSuffix(name, regTimestampSuffix)
}

func (r Cleaner2) BucketObjectNamesChan(
	ctx CancelableContext,
	bucketName string,
	bucketObjectChan chan BucketObject,
	errChan chan error) {

	r.iterateBucketObjectNames(ctx, bucketName, bucketObjectChan, errChan)
}

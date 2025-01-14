package gs

import (
	"context"

	"cloud.google.com/go/storage"

	"github.com/c2fo/vfs/v6"
)

// bucketHandle is an interface which contains a subset of the functions provided
// by storage.BucketHandler. Any function normally called directly by storage.BucketHandler
// should be added to this interface to allow for proper retry wrapping of the functions
// which call the GCS API.
type bucketHandle interface {
	Attrs(ctx context.Context) (*storage.BucketAttrs, error)
}

// bucketHandleWrapper is a unique, wrapped type which should mimic the behavior of BucketHandler, but with
// modified return types. Each function that returns a sub type that also should be wrapped should be added
// to this interface with the 'Wrapped' prefix.
type bucketHandleWrapper interface {
	bucketHandle
	WrappedObjects(ctx context.Context, q *storage.Query) objectIteratorWrapper
}

// retryBucketHandler implements the bucketHandle interface
type retryBucketHandler struct {
	Retry   vfs.Retry
	handler *storage.BucketHandle
}

// Attrs accepts a context and returns bucket attrs wrapped in a retry
func (r *retryBucketHandler) Attrs(ctx context.Context) (*storage.BucketAttrs, error) {
	return bucketAttributeRetry(r.Retry, func() (*storage.BucketAttrs, error) {
		return r.handler.Attrs(ctx)
	})
}

// WrappedObjects returns an iterator over the objects in the bucket that match the Query q, all wrapped in a retry.
// If q is nil, no filtering is done.
func (r *retryBucketHandler) WrappedObjects(ctx context.Context, q *storage.Query) objectIteratorWrapper {
	return &retryObjectIterator{Retry: r.Retry, iterator: r.handler.Objects(ctx, q)}
}

// objectIteratorWrapper is an interface which contains a subset of the functions provided by storage.ObjectIterator.
type objectIteratorWrapper interface {
	Next() (*storage.ObjectAttrs, error)
}

// retryObjectIterator implements the objectIteratorWrapper interface
type retryObjectIterator struct {
	Retry    vfs.Retry
	iterator *storage.ObjectIterator
}

// Next returns the next result, wrapped in retry. Its second return value is iterator.Done if
// there are no more results. Once Next returns iterator.Done, all subsequent
// calls will return iterator.Done.
//
// If Query.Delimiter is non-empty, some of the ObjectAttrs returned by Next will
// have a non-empty Prefix field, and a zero value for all other fields. These
// represent prefixes.
func (r *retryObjectIterator) Next() (*storage.ObjectAttrs, error) {
	return objectAttributeRetry(r.Retry, func() (*storage.ObjectAttrs, error) {
		return r.iterator.Next()
	})
}

func bucketAttributeRetry(retry vfs.Retry, attrFunc func() (*storage.BucketAttrs, error)) (*storage.BucketAttrs, error) {
	var attrs *storage.BucketAttrs
	if err := retry(func() error {
		var retryErr error
		attrs, retryErr = attrFunc()
		if retryErr != nil {
			return retryErr
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return attrs, nil
}

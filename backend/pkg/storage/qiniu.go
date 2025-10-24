package storage

type QiniuStorage struct {
	accessKey string
	secretKey string
	bucket    string
	domain    string
}

func NewQiniuStorage(accessKey, secretKey, bucket, domain string) *QiniuStorage {
	return &QiniuStorage{
		accessKey: accessKey,
		secretKey: secretKey,
		bucket:    bucket,
		domain:    domain,
	}
}

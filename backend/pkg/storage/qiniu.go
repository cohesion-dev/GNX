package storage

type QiniuClient struct {
	accessKey string
	secretKey string
	bucket    string
	domain    string
}

func NewQiniuClient(accessKey, secretKey, bucket, domain string) *QiniuClient {
	return &QiniuClient{
		accessKey: accessKey,
		secretKey: secretKey,
		bucket:    bucket,
		domain:    domain,
	}
}

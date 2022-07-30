package media

import "github.com/tsmweb/chasam/app/hash"

type Repository interface {
	AppendHash(hashType hash.Type, hashValue string, fileName string)
	FindByHash(hashType hash.Type, hashValue string) string
	AppendPerceptualHash(hashType hash.Type, hashValue uint64, fileName string)
	FindByPerceptualHash(hashType hash.Type, hashValue uint64, distance int) (int, string)
}

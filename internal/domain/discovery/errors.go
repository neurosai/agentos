package discovery

import "errors"

var (
	ErrUnsafeCollector        = errors.New("unsafe collector not allowed in v0.1")
	ErrInvalidMode            = errors.New("discovery mode must be read_only")
	ErrMemoryWithoutCatalog   = errors.New("writeToMemory requires writeToCatalog")
)

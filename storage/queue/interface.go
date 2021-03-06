package queue

import "yu/storage"

type Queue interface {
	storage.StorageType
	Push([]byte) error
	Pop() ([]byte, error)
}

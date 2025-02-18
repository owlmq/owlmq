package storage

type StorageType int

const (
	InMemory StorageType = iota
	File
)

func (s StorageType) String() string {
	switch s {
	case InMemory:
		return "InMemory"
	case File:
		return "File"
	default:
		return "Unknown"
	}
}

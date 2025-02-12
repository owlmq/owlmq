package repository

type Repository_layer interface {
	Write(key string, value string) (err error)
	Read(key string) (value string, err error)
}

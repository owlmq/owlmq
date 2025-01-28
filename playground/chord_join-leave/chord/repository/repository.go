package repository

import (
	"errors"
	"sync"
)

var lock = &sync.Mutex{}
var singleInstance *Repository

func New() *Repository {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()

		if singleInstance == nil {
			singleInstance = &Repository{
				//init map as storage
				store: make(map[string]string),
			}
		}
	}
	return singleInstance
}

type Repository struct {
	store map[string]string
}

func (r *Repository) Write(key string, value string) (err error) {
	r.store[key] = value
	return nil
}

func (r *Repository) Read(key string) (value string, err error) {
	val, ok := r.store[key]
	if !ok {
		//the string is empty so that if you use the function wrong there is no runtime error; normaly you would return a nil pointer but in this case like every repository layer should we try to prevent the system from runtime errors
		return "", errors.New(" key not found in map")
	}
	return val, nil
}

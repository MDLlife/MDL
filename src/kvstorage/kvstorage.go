package kvstorage

import (
	"errors"
	"sync"

	"github.com/MDLlife/MDL/src/util/file"
)

var (
	// ErrNoSuchKey is returned when the specified key does not exist
	// in the storage instance
	ErrNoSuchKey = NewError(errors.New("no such key exists in the storage"))
)

// kvStorage is a key-value storage for storing arbitrary data
type kvStorage struct {
	fileName string
	data     map[string]string
	sync.RWMutex
}

// newKVStorage constructs new storage instance using the file with the `fileName`
// to persist data
func newKVStorage(fileName string) (*kvStorage, error) {
	storage := kvStorage{
		fileName: fileName,
	}

	if err := file.LoadJSON(fileName, &storage.data); err != nil {
		return nil, err
	}

	return &storage, nil
}

// get gets the value associated with the `key`. Returns `ErrNoSuchKey`
func (s *kvStorage) get(key string) (string, error) {
	s.RLock()
	defer s.RUnlock()

	val, ok := s.data[key]
	if !ok {
		return "", ErrNoSuchKey
	}

	return val, nil
}

// getAll gets the snapshot of the current storage contents
func (s *kvStorage) getAll() map[string]string {
	s.RLock()
	defer s.RUnlock()

	return copyMap(s.data)
}

// add adds the `val` value to the storage with the specified `key`. Replaces the
// original value if `key` already exists
func (s *kvStorage) add(key, val string) error {
	s.Lock()
	defer s.Unlock()

	// save original data
	oldVal, oldOk := s.data[key]

	s.data[key] = val

	// try to persist data, fall back to original data on error
	if err := s.flush(); err != nil {
		if !oldOk {
			delete(s.data, key)
		} else {
			s.data[key] = oldVal
		}

		return err
	}

	return nil
}

// remove removes the value associated with the `key`. Returns `ErrNoSuchKey`
func (s *kvStorage) remove(key string) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.data[key]; !ok {
		return ErrNoSuchKey
	}

	// save original data
	oldVal := s.data[key]

	delete(s.data, key)

	// try to persist data, fall back to original data on error
	if err := s.flush(); err != nil {
		s.data[key] = oldVal

		return err
	}

	return nil
}

// flush persists data to file
func (s *kvStorage) flush() error {
	return file.SaveJSON(s.fileName, s.data, 0600)
}

package syncmap

import (
	"errors"
	"sync"
)

type SyncMapInterface interface {
	Get(key string) (chan []byte, error)
	Set(key string, value chan []byte) error
	Del(key string) error
}

type syncMap struct {
	keyChanelMap map[string]chan []byte
	mutex        *sync.RWMutex
}

func NewSyncMap() SyncMapInterface {
	return &syncMap{
		keyChanelMap: make(map[string]chan []byte),
		mutex:        &sync.RWMutex{},
	}
}

func (s *syncMap) Get(key string) (chan []byte, error) {
	s.mutex.RLock()
	c, ok := s.keyChanelMap[key]
	s.mutex.RUnlock()

	if !ok {
		return nil, errors.New("key don't found")
	}

	return c, nil
}

func (s *syncMap) Set(key string, value chan []byte) error {
	s.mutex.RLock()
	_, ok := s.keyChanelMap[key]
	s.mutex.RUnlock()

	if ok {
		return errors.New("key found")
	}

	s.mutex.Lock()
	s.keyChanelMap[key] = value
	s.mutex.Unlock()

	return nil
}

func (s *syncMap) Del(key string) error {
	s.mutex.RLock()
	_, ok := s.keyChanelMap[key]
	s.mutex.RUnlock()

	if !ok {
		return errors.New("key don't found")
	}

	s.mutex.Lock()
	delete(s.keyChanelMap, key)
	s.mutex.Unlock()

	return nil
}

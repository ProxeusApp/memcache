package cache

import (
	"time"
	"sync"
	"reflect"
	"os"
)

var secondsAfter time.Duration = 1

type value struct{
	expiry time.Time
	access time.Time
	val    interface{}
}

type Cache struct {
	store             map[interface{}]*value
	Expiry            time.Duration
	cleanupTimer      *time.Timer
	cleanupLock       sync.Mutex
	cacheLock         sync.RWMutex
	OnExpired	  func(key interface{}, val interface{})
}

func NewCache(expiry time.Duration) *Cache{
	c := &Cache{store:make(map[interface{}]*value)}
	c.Expiry = expiry
	return c
}

func (s *Cache) Get(key interface{}, ref interface{}) error{
	s.cacheLock.RLock()
	valueHolder := s.store[key]
	s.cacheLock.RUnlock()

	if valueHolder != nil {
		v := reflect.ValueOf(ref)
		if v.Kind() != reflect.Ptr || v.IsNil() {
			return os.ErrInvalid
		}

		//update last touch
		n := time.Now()
		s.cacheLock.Lock()
		valueHolder.access = n
		valueHolder.expiry = valueHolder.access.Add(s.Expiry)
		s.cacheLock.Lock()

		i := 0
		for v.Kind() != reflect.Struct && v.Kind() != reflect.Invalid && (!v.CanSet() || v.Type() != reflect.TypeOf(valueHolder.val)) {
			v = v.Elem()
			if i > 3 {
				break
			}
			i++
		}
		if !v.CanSet() || v.Kind() != reflect.TypeOf(valueHolder.val).Kind() {
			return os.ErrInvalid
		}
		v.Set(reflect.ValueOf(valueHolder.val))
		return nil
	}
	return os.ErrNotExist
}

func (s *Cache) Remove(key interface{}) bool{
	s.cacheLock.Lock()
	session := s.store[key]
	if session != nil {
		delete(s.store, key)
		s.cacheLock.Unlock()
		return true
	}else{
		s.cacheLock.Unlock()
		return false
	}
}

func (s *Cache) PutWithOtherExpiry(key interface{}, val interface{}, expiry time.Duration) {
	n := time.Now()
	exp := n.Add(expiry)
	session := &value{expiry:exp, access:n, val:val}
	s.cacheLock.Lock()
	s.store[key]=session
	s.cacheLock.Unlock()
	s.startCleanup(expiry+(secondsAfter*time.Second))
}


func (s *Cache) Put(key interface{}, val interface{}) {
	n := time.Now()
	expiry := n.Add(s.Expiry)
	session := &value{expiry:expiry, access:n, val:val}
	s.cacheLock.Lock()
	s.store[key]=session
	s.cacheLock.Unlock()
	s.startCleanup(s.Expiry+(secondsAfter*time.Second))
}

func (s *Cache) cleanupScheduler(){
	n := time.Now()
	s.cleanupLock.Lock()
	var minExpiry time.Time = n.Add(s.Expiry)
	expiredSessions := make(map[interface{}]*value)
	s.cacheLock.Lock()
	var key interface{}
	var val *value
	for key = range s.store {
		val = s.store[key]
		if n.After(val.expiry){
			expiredSessions[key] = val
		}else if val.expiry.Before(minExpiry){
			minExpiry = val.expiry
		}
	}
	for key = range expiredSessions{
		if s.OnExpired != nil {
			s.OnExpired(key, expiredSessions[key].val);
		}
		delete(s.store, key)
	}
	s.cacheLock.Unlock()
	for key = range expiredSessions{
		val = expiredSessions[key]
	}
	sessionsLength := len(s.store)
	if sessionsLength > 0 {
		nextRunIn := minExpiry.Sub(n)+(secondsAfter*time.Second)
		s.cleanupTimer = time.AfterFunc(nextRunIn, s.cleanupScheduler)
	}else{
		s.cleanupTimer = nil
	}
	s.cleanupLock.Unlock()
}

func (s *Cache) startCleanup(runAfter time.Duration){
	if s.cleanupTimer == nil {
		s.cleanupLock.Lock()
		if s.cleanupTimer == nil {
			s.cleanupTimer = time.AfterFunc(runAfter, s.cleanupScheduler)
		}
		s.cleanupLock.Unlock()
	}
}

func (s *Cache) stopCleanup(){
	s.cleanupLock.Lock()
	if s.cleanupTimer != nil {
		s.cleanupTimer.Stop()
		s.cleanupTimer = nil
	}
	s.cleanupLock.Unlock()
}

func (s *Cache) Close(){
	s.stopCleanup()
}


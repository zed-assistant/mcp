package instance

import (
	"sync"
)

type InstanceLockManager struct {
	locks sync.Map
}

var (
	lockManagerSingleton InstanceLockManager
	lockManagerOnce      sync.Once
)

func NewInstanceLockManager() *InstanceLockManager {
	lockManagerOnce.Do(func() {
		lockManagerSingleton = InstanceLockManager{
			locks: sync.Map{},
		}
	})
	return &lockManagerSingleton
}

func (m *InstanceLockManager) Lock(instanceId string) {
	stored, _ := m.locks.LoadOrStore(instanceId, &sync.RWMutex{})
	if lock, ok := stored.(*sync.RWMutex); ok {
		lock.Lock()
	}
}

func (m *InstanceLockManager) Unlock(instanceId string) {
	if stored, exists := m.locks.Load(instanceId); exists {
		if lock, ok := stored.(*sync.RWMutex); ok {
			lock.Unlock()
		}
	}
}

func (m *InstanceLockManager) RLock(instanceId string) {
	stored, _ := m.locks.LoadOrStore(instanceId, &sync.RWMutex{})
	if lock, ok := stored.(*sync.RWMutex); ok {
		lock.RLock()
	}
}

func (m *InstanceLockManager) RUnlock(instanceId string) {
	if stored, exists := m.locks.Load(instanceId); exists {
		if lock, ok := stored.(*sync.RWMutex); ok {
			lock.RUnlock()
		}
	}
}

func (m *InstanceLockManager) CleanupLock(instanceId string) {
	m.locks.Delete(instanceId)
}

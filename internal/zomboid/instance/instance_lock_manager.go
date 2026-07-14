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
	lock, _ := m.locks.LoadOrStore(instanceId, &sync.RWMutex{})
	lock.(*sync.RWMutex).Lock()
}

func (m *InstanceLockManager) Unlock(instanceId string) {
	if lock, exists := m.locks.Load(instanceId); exists {
		lock.(*sync.RWMutex).Unlock()
	}
}

func (m *InstanceLockManager) RLock(instanceId string) {
	lock, _ := m.locks.LoadOrStore(instanceId, &sync.RWMutex{})
	lock.(*sync.RWMutex).RLock()
}

func (m *InstanceLockManager) RUnlock(instanceId string) {
	if lock, exists := m.locks.Load(instanceId); exists {
		lock.(*sync.RWMutex).RUnlock()
	}
}

func (m *InstanceLockManager) CleanupLock(instanceId string) {
	m.locks.Delete(instanceId)
}

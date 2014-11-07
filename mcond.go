package mcond

import (
	"fmt"
	"net/http"
	"sync"

	redis "gopkg.in/redis.v2"
)

type MCond struct {
	conds map[string]*sync.Cond
	hosts []string
	redis *redis.Client
}

func NewMCond(client *redis.Client) (m *MCond) {
	return &MCond{
		conds: map[string]*sync.Cond{},
		redis: client,
		hosts: []string{},
	}
}

func (m *MCond) AddHost(host string) {
	for _, item := range m.hosts {
		if item == host {
			return
		}
	}
	m.hosts = append(m.hosts, host)
}

func (m *MCond) AddCond(key string) {
	if m.conds[key] != nil {
		return
	}
	m.conds[key] = sync.NewCond(&sync.Mutex{})
}

func (m *MCond) cond(key string) (c *sync.Cond) {
	c = m.conds[key]
	return
}

func (m *MCond) Wait(key string) {
	if m.conds[key] != nil {
		m.conds[key].Wait()
	}
}

func (m *MCond) BroadcastSelf(key string) {
	if m.cond(key) != nil {
		m.cond(key).Broadcast()
	}
}

func (m *MCond) broadcastPath(host, key string) string {
	return fmt.Sprintf("http://%s/cache/broadcast/%s", host, key)
}

func (m *MCond) Broadcast(key string) {
	m.BroadcastSelf(key)
	for _, item := range m.hosts {
		http.Get(m.broadcastPath(item, key))
	}
}

func (m *MCond) processingKey(key string) string {
	return fmt.Sprintf("mcond:processing:%s", key)
}

func (m *MCond) completeKey(key string) string {
	return fmt.Sprintf("mcond:complete:%s", key)
}

func (m *MCond) AddProcessing(key string) {
	m.redis.Set(m.processingKey(key), "1").Result()
}

func (m *MCond) AddCompleted(key string) {
	m.redis.Del(m.processingKey(key)).Result()
	m.redis.Set(m.completeKey(key), "1").Result()
}

func (m *MCond) IsProcessing(key string) bool {
	result, _ := m.redis.Get(m.processingKey(key)).Result()
	return result == "1"
}

func (m *MCond) IsComplete(key string) bool {
	result, _ := m.redis.Get(m.completeKey(key)).Result()
	return result == "1"
}

func (m *MCond) WaitForAvailable(key string) {
	m.cond(key).L.Lock()
	for m.IsProcessing(key) || !m.IsComplete(key) {
		m.Wait(key)
	}
	m.cond(key).L.Unlock()
}

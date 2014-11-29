package mcond

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	redis "gopkg.in/redis.v2"
)

const (
	bloadcastPath = "/cache/broadcast/"
)

type MCond struct {
	conds map[string]*sync.Cond
	hosts []string
	redis *redis.Client
	addr  string
}

type MCondOption struct {
	RedisHost string
	RedisDB   int64
	MCondAddr string
}

func NewMCond(option MCondOption) (m *MCond) {
	if option.RedisHost == "" {
		option.RedisHost = "localhost:6379"
	}
	if option.MCondAddr == "" {
		option.MCondAddr = ":9012"
	}
	return &MCond{
		conds: map[string]*sync.Cond{},
		redis: redis.NewTCPClient(&redis.Options{
			Addr: option.RedisHost,
			DB:   option.RedisDB,
		}),
		addr:  option.MCondAddr,
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
	return fmt.Sprintf("http://%s%s%s", host, bloadcastPath, key)
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

func (m *MCond) Clear() {
	for key := range m.conds {
		m.redis.Del(m.processingKey(key)).Result()
		m.redis.Del(m.completeKey(key)).Result()
	}
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

func (m MCond) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if !strings.HasPrefix(path, bloadcastPath) {
		return
	}
	key := strings.Replace(path, bloadcastPath, "", 1)
	m.Broadcast(key)
}

func (m *MCond) Start() {
	s := &http.Server{
		Addr:           m.addr,
		Handler:        m,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		log.Fatal(s.ListenAndServe())
	}()
}

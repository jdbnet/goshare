package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"goshare/config"
	"goshare/logger"
)

type FileMeta struct {
	ID           string    `json:"id"`
	OriginalName string    `json:"original_name"`
	Size         int64     `json:"size"`
	UploadedAt   time.Time `json:"uploaded_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Manager struct {
	cfg        *config.Config
	metaFile   string
	metaMutex  sync.RWMutex
	files      map[string]FileMeta
	stopChan   chan struct{}
}

func NewManager(cfg *config.Config) (*Manager, error) {
	if err := os.MkdirAll(cfg.Storage.Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage dir: %v", err)
	}

	m := &Manager{
		cfg:      cfg,
		metaFile: filepath.Join(cfg.Storage.Dir, "metadata.json"),
		files:    make(map[string]FileMeta),
		stopChan: make(chan struct{}),
	}

	if err := m.loadMetadata(); err != nil {
		logger.Warn("Failed to load metadata, starting fresh: %v", err)
	}

	return m, nil
}

func (m *Manager) loadMetadata() error {
	m.metaMutex.Lock()
	defer m.metaMutex.Unlock()

	data, err := os.ReadFile(m.metaFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, &m.files); err != nil {
		return err
	}
	return nil
}

func (m *Manager) saveMetadata() error {
	// Not taking lock here, caller must hold it
	data, err := json.MarshalIndent(m.files, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.metaFile, data, 0644)
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (m *Manager) SaveFile(originalName string, r io.Reader, size int64, expiresAt time.Time) (string, error) {
	id := generateID()
	destPath := filepath.Join(m.cfg.Storage.Dir, id)

	f, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		os.Remove(destPath)
		return "", err
	}

	meta := FileMeta{
		ID:           id,
		OriginalName: originalName,
		Size:         size,
		UploadedAt:   time.Now(),
		ExpiresAt:    expiresAt,
	}

	m.metaMutex.Lock()
	defer m.metaMutex.Unlock()

	m.files[id] = meta
	if err := m.saveMetadata(); err != nil {
		logger.Error("Failed to save metadata after upload: %v", err)
	}

	return id, nil
}

func (m *Manager) GetFileMeta(id string) (FileMeta, bool) {
	m.metaMutex.RLock()
	defer m.metaMutex.RUnlock()
	meta, ok := m.files[id]
	return meta, ok
}

func (m *Manager) GetFilePath(id string) string {
	return filepath.Join(m.cfg.Storage.Dir, id)
}

func (m *Manager) GetAllFiles() []FileMeta {
	m.metaMutex.RLock()
	defer m.metaMutex.RUnlock()

	var result []FileMeta
	for _, f := range m.files {
		result = append(result, f)
	}
	return result
}

func (m *Manager) DeleteFile(id string) error {
	m.metaMutex.Lock()
	defer m.metaMutex.Unlock()

	if _, ok := m.files[id]; !ok {
		return fmt.Errorf("file not found")
	}

	path := filepath.Join(m.cfg.Storage.Dir, id)
	os.Remove(path) // Ignore error if file is already gone from disk

	delete(m.files, id)
	return m.saveMetadata()
}

func (m *Manager) StartCleanupTask() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				m.cleanupExpired()
			case <-m.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (m *Manager) StopCleanupTask() {
	close(m.stopChan)
}

func (m *Manager) cleanupExpired() {
	m.metaMutex.Lock()
	defer m.metaMutex.Unlock()

	now := time.Now()
	changed := false

	for id, meta := range m.files {
		if now.After(meta.ExpiresAt) {
			logger.Info("File expired, deleting: %s (%s)", meta.OriginalName, id)
			path := filepath.Join(m.cfg.Storage.Dir, id)
			os.Remove(path)
			delete(m.files, id)
			changed = true
		}
	}

	if changed {
		if err := m.saveMetadata(); err != nil {
			logger.Error("Failed to save metadata after cleanup: %v", err)
		}
	}
}

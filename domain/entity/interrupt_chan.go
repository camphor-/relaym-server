package entity

import (
	"fmt"
	"sync"

	"github.com/camphor-/relaym-server/log"
)

// SyncCheckTimer はSpotifyとの同期チェック用のタイマーです。タイマーが止まったことを確認するためのstopチャネルがあります。
// ref : http://okzk.hatenablog.com/entry/2015/12/01/001924
type InterruptCheckChan struct {
	interruptCh chan struct{}
}

// InterruptCh は次の曲への遷移の指示を送るチャネルを返します。
func (s *InterruptCheckChan) InterruptCh() <-chan struct{} {
	return s.interruptCh
}

func newInterruptCh() *InterruptCheckChan {
	return &InterruptCheckChan{
		interruptCh: make(chan struct{}, 1),
	}
}

// SyncCheckTimerManager はSpotifyとの同期チェック用のタイマーを一括して管理する構造体です。
type InterruptChanManager struct {
	chans map[string]*InterruptCheckChan
	mu    sync.Mutex
}

// NewInterruptChanManager はInterruptChanManagerのポインタを生成します。
func NewInterruptChanManager() *InterruptChanManager {
	return &InterruptChanManager{
		chans: map[string]*InterruptCheckChan{},
	}
}

// CreateInterruptCheckChan は与えられたセッションの同期チェック用のタイマーを作成します。
// 既存のタイマーが存在する場合はstopしてから新しいタイマーを作成します。
func (m *InterruptChanManager) CreateInterruptCheckChan(sessionID string) *InterruptCheckChan {
	logger := log.New()
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Debugj(map[string]interface{}{"message": "create InterruptCheckChan", "sessionID": sessionID})

	if existing, ok := m.chans[sessionID]; ok {
		logger.Debugj(map[string]interface{}{"message": "InterruptCheckChan has already exists", "sessionID": sessionID})
		close(existing.interruptCh)
	}
	chans := newInterruptCh()
	m.chans[sessionID] = chans
	return chans
}

// GetChan は与えられたセッションのタイマーを取得します。存在しない場合はfalseが返ります。
func (m *InterruptChanManager) GetChan(sessionID string) (*InterruptCheckChan, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.chans[sessionID]; ok {
		return existing, true
	}
	return nil, false
}

func (m *InterruptChanManager) DeleteChan(sessionID string) {
	logger := log.New()
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Debugj(map[string]interface{}{"message": "delete InterruptCheckChan", "sessionID": sessionID})

	if timer, ok := m.chans[sessionID]; ok {
		close(timer.interruptCh)
		delete(m.chans, sessionID)
		return
	}

	logger.Debugj(map[string]interface{}{"message": "InterruptCheckChan not existed", "sessionID": sessionID})
}

func (m *InterruptChanManager) InterruptChan(sessionID string) error {
	logger := log.New()
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Debugj(map[string]interface{}{"message": "stop InterruptCheckChan", "sessionID": sessionID})

	if timer, ok := m.chans[sessionID]; ok {
		timer.interruptCh <- struct{}{}
		return nil
	}

	logger.Debugj(map[string]interface{}{"message": "InterruptCheckChan not existed", "sessionID": sessionID})
	return fmt.Errorf("InterruptCheckChan not existed")
}

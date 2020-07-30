package entity

import (
	"sync"
	"time"

	"github.com/camphor-/relaym-server/log"
)

// SyncCheckTimer はSpotifyとの同期チェック用のタイマーです。タイマーが止まったことを確認するためのstopチャネルがあります。
// ref : http://okzk.hatenablog.com/entry/2015/12/01/001924
type SyncCheckTimer struct {
	timer  *time.Timer
	stopCh chan struct{}
	nextCh chan string
}

// ExpireCh は指定設定された秒数経過したことを送るチャネルを返します。
func (s *SyncCheckTimer) ExpireCh() <-chan time.Time {
	return s.timer.C
}

// StopCh はタイマーがストップされたことを送るチャネルを返します。
func (s *SyncCheckTimer) StopCh() <-chan struct{} {
	return s.stopCh
}

// NextCh は次の曲への遷移の指示を送るチャネルを返します。
func (s *SyncCheckTimer) NextCh() <-chan string {
	return s.nextCh
}

func newSyncCheckTimer(d time.Duration) *SyncCheckTimer {
	return &SyncCheckTimer{
		timer:  time.NewTimer(d),
		stopCh: make(chan struct{}, 2),
		nextCh: make(chan string, 1),
	}
}

// SyncCheckTimerManager はSpotifyとの同期チェック用のタイマーを一括して管理する構造体です。
type SyncCheckTimerManager struct {
	timers map[string]*SyncCheckTimer
	mu     sync.Mutex
}

// NewSyncCheckTimerManager はSyncCheckTimerManagerのポインタを生成します。
func NewSyncCheckTimerManager() *SyncCheckTimerManager {
	return &SyncCheckTimerManager{
		timers: map[string]*SyncCheckTimer{},
	}
}

// CreateTimer は与えられたセッションの同期チェック用のタイマーを作成します。
// 既存のタイマーが存在する場合はstopしてから新しいタイマーを作成します。
func (m *SyncCheckTimerManager) CreateTimer(sessionID string, d time.Duration) *SyncCheckTimer {
	logger := log.New()
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Debugj(map[string]interface{}{"message": "create timer", "sessionID": sessionID})

	if existing, ok := m.timers[sessionID]; ok {
		// 本来ならStopのGoDocコメントにある通り、<-t.Cとして、チャネルが空になっていることを確認すべきだが、
		// ExpireCh()の呼び出し側で受け取っているので問題ない。
		logger.Debugj(map[string]interface{}{"message": "timer has already exists", "sessionID": sessionID})
		existing.timer.Stop()
		close(existing.stopCh)
	}
	timer := newSyncCheckTimer(d)
	m.timers[sessionID] = timer
	return timer
}

// StopTimer は与えられたセッションのタイマーを終了します。
func (m *SyncCheckTimerManager) StopTimer(sessionID string) {
	logger := log.New()
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Debugj(map[string]interface{}{"message": "stop timer", "sessionID": sessionID})

	if timer, ok := m.timers[sessionID]; ok {
		if !timer.timer.Stop() {
			<-timer.timer.C
		}
		close(timer.stopCh)
		delete(m.timers, sessionID)
		return
	}

	logger.Debugj(map[string]interface{}{"message": "timer not existed", "sessionID": sessionID})
}

// DeleteTimer は与えられたセッションのタイマーをマップから削除します。
// StopTimerと異なり、タイマーのストップ処理は行いません。
// 既にタイマーがExpireして、そのチャネルの値を取り出してしまった後にマップから削除したいときに使います。
// <-timer.timer.Cを呼ぶと無限に待ちが発生してしまいます。(値を取り出すことは一生出来ないので)
func (m *SyncCheckTimerManager) DeleteTimer(sessionID string) {
	logger := log.New()
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Debugj(map[string]interface{}{"message": "delete timer", "sessionID": sessionID})

	if timer, ok := m.timers[sessionID]; ok {
		close(timer.stopCh)
		delete(m.timers, sessionID)
		return
	}

	logger.Debugj(map[string]interface{}{"message": "timer not existed", "sessionID": sessionID})
}

// GetTimer は与えられたセッションのタイマーを取得します。存在しない場合はfalseが返ります。
func (m *SyncCheckTimerManager) GetTimer(sessionID string) (*SyncCheckTimer, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.timers[sessionID]; ok {
		return existing, true
	}
	return nil, false
}

// ExpireTimer は与えられたセッションのタイマーをExpiredさせます。
func (m *SyncCheckTimerManager) ExpireTimer(sessionID string) {
	logger := log.New()
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Debugj(map[string]interface{}{"message": "stop timer", "sessionID": sessionID})

	if timer, ok := m.timers[sessionID]; ok {
		timer.nextCh <- "next track"
		return
	}

	logger.Debugj(map[string]interface{}{"message": "timer not existed", "sessionID": sessionID})
}

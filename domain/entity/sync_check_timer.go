package entity

import (
	"fmt"
	"sync"
	"time"

	"github.com/camphor-/relaym-server/log"
)

// SyncCheckTimer はSpotifyとの同期チェック用のタイマーです。タイマーが止まったことを確認するためのstopチャネルがあります。
// ref : http://okzk.hatenablog.com/entry/2015/12/01/001924
type SyncCheckTimer struct {
	timer          *time.Timer
	isTimerExpired bool
	stopCh         chan struct{}
	nextCh         chan struct{}
	mutexForNextCh sync.Mutex
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
func (s *SyncCheckTimer) NextCh() <-chan struct{} {
	return s.nextCh
}

func (s *SyncCheckTimer) LockNextCh() {
	logger := log.New()
	s.mutexForNextCh.Lock()
	logger.Debugj(map[string]interface{}{"message": "next ch locked success"})
}

func (s *SyncCheckTimer) UnlockNextCh() {
	logger := log.New()
	s.mutexForNextCh.Unlock()
	logger.Debugj(map[string]interface{}{"message": "next ch unlocked success"})
	// UnlockがUnlock待ちになってるlockに通知されるのを待つ
	time.Sleep(10 * time.Millisecond)
}

// closeNextCh はNextChを安全にcloseします
func (s *SyncCheckTimer) closeNextCh() {
	for {
		select {
		case <-s.nextCh:
		default:
			close(s.nextCh)
			return
		}
	}
}

// MakeIsTimerExpiredTrue はisTimerExpiredをtrueに変更します
// <- s.ExpireCh でtimerから値を受け取った際に呼び出してください
func (s *SyncCheckTimer) MakeIsTimerExpiredTrue() {
	s.isTimerExpired = true
}

// newSyncCheckTimer はSyncCheckTimerを作成します
// この段階ではtimerには空のtimerがセットされており、SetTimerを使用して正しいtimerのセットを行う必要があります
func newSyncCheckTimer() *SyncCheckTimer {
	timer := time.NewTimer(0)
	//Expiredしたtimerを作成する
	if !timer.Stop() {
		<-timer.C
	}

	return &SyncCheckTimer{
		stopCh:         make(chan struct{}, 2),
		nextCh:         make(chan struct{}, 1),
		isTimerExpired: true,
		timer:          timer,
	}
}

// SetTimerはSyncCheckTimerにTimerをセットします
func (s *SyncCheckTimer) SetDuration(d time.Duration) {
	if !s.timer.Stop() && !s.isTimerExpired {
		<-s.timer.C
	}

	s.isTimerExpired = false
	s.timer.Reset(d)
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

// CreateExpiredTimer は与えられたセッションの同期チェック用のタイマーを作成します。
// 既存のタイマーが存在する場合はstopしてから新しいタイマーを作成します。
func (m *SyncCheckTimerManager) CreateExpiredTimer(sessionID string) *SyncCheckTimer {
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
	timer := newSyncCheckTimer()
	m.timers[sessionID] = timer
	return timer
}

// DeleteTimer は与えられたセッションのタイマーをマップから削除します。
// 既にタイマーがExpireして、そのチャネルの値を取り出してしまった後にマップから削除したいときに使います。
func (m *SyncCheckTimerManager) DeleteTimer(sessionID string) {
	logger := log.New()
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Debugj(map[string]interface{}{"message": "delete timer", "sessionID": sessionID})

	if timer, ok := m.timers[sessionID]; ok {
		close(timer.stopCh)
		timer.closeNextCh()
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

// SendToNextCh は与えられたセッションのタイマーのNextChに通知を送ります
// goroutineで実行されることを想定しています
func (m *SyncCheckTimerManager) SendToNextCh(sessionID string) {
	logger := log.New()

	logger.Debugj(map[string]interface{}{"message": "call next ch", "sessionID": sessionID})

	if timer, ok := m.timers[sessionID]; ok {
		timer.LockNextCh()
		select {
		case timer.nextCh <- struct{}{}:
			logger.Debugj(map[string]interface{}{"message": "nextCh unlocked and send to nextCh", "sessionID": sessionID})
			return
		default:
			logger.Debugj(map[string]interface{}{"message": "nextCh unlocked but can't send to nextCh", "sessionID": sessionID})
			return
		}
	}

	logger.Debugj(map[string]interface{}{"message": "timer not existed on SendToNextCh", "sessionID": sessionID})
}

// IsTimerExpired は与えられたセッションのisTimerExpiredの値を返します
func (m *SyncCheckTimerManager) IsTimerExpired(sessionID string) (bool, error) {
	logger := log.New()
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.timers[sessionID]; ok {
		return existing.isTimerExpired, nil
	}

	logger.Debugj(map[string]interface{}{"message": "timer not existed on IsRemainDuration", "sessionID": sessionID})
	return false, fmt.Errorf("timer not existed")
}

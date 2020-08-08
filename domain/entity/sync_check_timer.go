package entity

import (
	"fmt"
	"sync"
	"time"

	"github.com/camphor-/relaym-server/log"
)

// SyncCheckTimer はSpotifyとの同期チェック用のタイマーです。タイマーが止まったことを確認するためのstopチャネルがあります。
// ref : http://okzk.hatenablog.com/entry/2015/12/01/001924
// nextChはisNextChEnableがtrueの時にのみ値を送ることを想定しています
type SyncCheckTimer struct {
	timer          *time.Timer
	isTimerExpired bool
	stopCh         chan struct{}
	isNextChEnable bool
	nextCh         chan struct{}
	enableNextCh   chan struct{}
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

// EnableNextCh はNextChを有効化します・
// isNextChEnableをtrueにセットし、enableNextChに通知します
func (s *SyncCheckTimer) EnableNextCh() {
	select {
	case s.enableNextCh <- struct{}{}:
		s.isNextChEnable = true
		return
	default:
		s.isNextChEnable = true
		return
	}
}

// DisableNextCh はNextChを無効化します。
// isNextChEnableをfalseにセットし、enableNextChの中身を0にします
func (s *SyncCheckTimer) DisableNextCh() {
	select {
	case <-s.enableNextCh:
		s.isNextChEnable = false
		return
	default:
		s.isNextChEnable = false
		return
	}
}

func (s *SyncCheckTimer) closeNextCh() {
	for {
		select {
		case <-s.nextCh:
			time.Sleep(1 * time.Millisecond)
		default:
			close(s.nextCh)
			return
		}
	}
}

// MakeIsTimerExpiredTrue はisTimerExpiredをtrueに変更します
// また、timerが確実にExpiredしていることを保証します、これの呼び出し時にExpiredしてないtimerはexpiredされます
// <- s.ExpireCh でtimerから値を受け取った際に呼び出してください
func (s *SyncCheckTimer) MakeIsTimerExpiredTrue() {
	s.SetDuration(0)
	if !s.timer.Stop() {
		<-s.timer.C
	}
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
		enableNextCh:   make(chan struct{}, 1),
		isTimerExpired: true,
		isNextChEnable: true,
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
		close(timer.enableNextCh)
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
// 大量にskipが叩かれた場合の待ちが予想されるので、goroutineで実行されることを想定しています
func (m *SyncCheckTimerManager) SendToNextCh(sessionID string) {
	logger := log.New()

	logger.Debugj(map[string]interface{}{"message": "call next ch", "sessionID": sessionID})

	if timer, ok := m.timers[sessionID]; ok {
		if timer.isNextChEnable {
			logger.Debugj(map[string]interface{}{"message": "timer existed and NextCh enable", "sessionID": sessionID})
			m.mu.Lock()
			timer.DisableNextCh()
			timer.nextCh <- struct{}{}
			m.mu.Unlock()
			return
		}
		logger.Debugj(map[string]interface{}{"message": "timer existed and NextCh disable", "sessionID": sessionID})
		// isNextChEnableがfalseの時はskip処理の途中なので、再度isNextChEnableになるまで待つ
		for {
			select {
			case _, ok := <-timer.enableNextCh:
				logger.Debugj(map[string]interface{}{"message": "timer NextCh be enable", "sessionID": sessionID})
				if ok {
					m.mu.Lock()
					timer.DisableNextCh()
					timer.nextCh <- struct{}{}
					m.mu.Unlock()
					return
				}
				return
			}
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

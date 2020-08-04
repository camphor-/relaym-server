package entity

import (
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/google/go-cmp/cmp"
)

func TestSyncCheckTimer_ExpireCh(t *testing.T) {
	t.Parallel()

	timer := time.NewTimer(10 * time.Second)

	type fields struct {
		timer  *time.Timer
		stopCh chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
		want   <-chan time.Time
	}{
		{
			name: "正しくチャネルを取得できる",
			fields: fields{
				timer:  timer,
				stopCh: nil,
			},
			want: timer.C,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SyncCheckTimer{
				timer:  tt.fields.timer,
				stopCh: tt.fields.stopCh,
			}
			if got := s.ExpireCh(); !cmp.Equal(got, tt.want) {
				t.Errorf("ExpireCh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncCheckTimer_StopCh(t *testing.T) {
	t.Parallel()

	stopCh := make(chan struct{}, 1)

	type fields struct {
		timer  *time.Timer
		stopCh chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
		want   <-chan struct{}
	}{
		{
			name: "正しくチャネルを取得できる",
			fields: fields{
				timer:  nil,
				stopCh: stopCh,
			},
			want: stopCh,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SyncCheckTimer{
				timer:  tt.fields.timer,
				stopCh: tt.fields.stopCh,
			}
			if got := s.StopCh(); !cmp.Equal(got, tt.want) {
				t.Errorf("StopCh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncCheckTimerManager_CreateExpiredTimer(t *testing.T) {
	t.Parallel()

	timer := newSyncCheckTimer()
	timer.SetDuration(time.Second)

	tests := []struct {
		name      string
		timers    map[string]*SyncCheckTimer
		sessionID string
		d         time.Duration
		want      *SyncCheckTimer
		ignoreCmp bool
	}{
		{
			name:      "まだタイマーが存在しないセッションのタイマーを作成できる",
			timers:    map[string]*SyncCheckTimer{},
			sessionID: "sessionID",
			d:         time.Second,
			want: &SyncCheckTimer{
				timer:  time.NewTimer(time.Second),
				stopCh: make(chan struct{}, 1),
			},
			ignoreCmp: true,
		},
		{
			name:      "すでにタイマーが存在するときでも新規のタイマーを返す",
			timers:    map[string]*SyncCheckTimer{"sessionID": timer},
			sessionID: "sessionID",
			d:         time.Second,
			want: &SyncCheckTimer{
				timer:  time.NewTimer(time.Second),
				stopCh: make(chan struct{}, 1),
			},
			ignoreCmp: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SyncCheckTimerManager{
				timers: tt.timers,
				mu:     sync.Mutex{},
			}

			if tt.ignoreCmp {
				return
			}
			opts := []cmp.Option{cmp.AllowUnexported(SyncCheckTimer{}), cmpopts.IgnoreUnexported(time.Timer{})}
			got := m.CreateExpiredTimer(tt.sessionID)
			got.SetDuration(tt.d)
			if !cmp.Equal(got, tt.want, opts...) {
				t.Errorf("CreateExpiredTimer() diff=%v", cmp.Diff(tt.want, got, opts...))
			}
		})
	}
}

func TestSyncCheckTimerManager_GetTimer(t *testing.T) {
	t.Parallel()

	session1Timer := time.NewTimer(10 * time.Second)

	tests := []struct {
		name      string
		timers    map[string]*SyncCheckTimer
		sessionID string
		want      *SyncCheckTimer
		want1     bool
	}{
		{
			name: "存在するセッションのタイマーを取得できる",
			timers: map[string]*SyncCheckTimer{"session1": {
				timer:  session1Timer,
				stopCh: nil,
			}},
			sessionID: "session1",
			want: &SyncCheckTimer{
				timer:  session1Timer,
				stopCh: nil,
			},
			want1: true,
		},
		{
			name: "存在しないセッションのタイマーのときはfalse",
			timers: map[string]*SyncCheckTimer{"session1": {
				timer:  session1Timer,
				stopCh: nil,
			}},
			sessionID: "not found session id",
			want:      nil,
			want1:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &SyncCheckTimerManager{
				timers: tt.timers,
				mu:     sync.Mutex{},
			}
			got, got1 := m.GetTimer(tt.sessionID)

			opts := []cmp.Option{cmp.AllowUnexported(SyncCheckTimer{}), cmpopts.IgnoreUnexported(time.Timer{})}
			if !cmp.Equal(got, tt.want, opts...) {
				t.Errorf("GetTimer() diff=%v", cmp.Diff(tt.want, got, opts...))
			}
			if got1 != tt.want1 {
				t.Errorf("GetTimer() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

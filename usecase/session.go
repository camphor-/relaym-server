package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/domain/spotify"
)

var syncCheckOffset = 5 * time.Second

// SessionUseCase はセッションに関するユースケースです。
type SessionUseCase struct {
	tm        *entity.SyncCheckTimerManager
	pusher    event.Pusher
	playerCli spotify.Player
}

// NewSessionUseCase はSessionUseCaseのポインタを生成します。
func NewSessionUseCase(playerCli spotify.Player, pusher event.Pusher) *SessionUseCase {
	return &SessionUseCase{
		tm:        entity.NewSyncCheckTimerManager(),
		pusher:    pusher,
		playerCli: playerCli,
	}
}

// ChangePlaybackState は与えられたセッションの再生状態を操作します。
func (s *SessionUseCase) ChangePlaybackState(ctx context.Context, sessionID string, st entity.StateType) error {
	switch st {
	case entity.Play:
		if err := s.play(ctx, sessionID); err != nil {
			return fmt.Errorf("play sessionID=%s: %w", sessionID, err)
		}
	case entity.Pause:
		if err := s.pause(ctx, sessionID); err != nil {
			return fmt.Errorf("pause sessionID=%s: %w", sessionID, err)
		}
	}
	return nil
}

// Play はセッションのstateを STOP → PLAY に変更して曲の再生を始めます。
func (s *SessionUseCase) play(ctx context.Context, sessionID string) error {
	// TODO : セッションを取得
	// sess ,err := repo.GetByID(sessionID)

	// TODO : デバイスIDをどっかから読み込む
	if err := s.playerCli.Play(ctx, ""); err != nil {
		return fmt.Errorf("call play api: %w", err)
	}

	// TODO : セッションのステートを書き換え
	// err := sess.MoveToPlay()
	// err := repo.Store(sess)

	go s.startSyncCheck(ctx, sessionID)

	s.pusher.Push(&event.PushMessage{
		SessionID: sessionID,
		Msg:       entity.EventPlay,
	})

	return nil
}

// Pause はセッションのstateをPLAY→PAUSEに変更して曲の再生を一時停止します。
func (s *SessionUseCase) pause(ctx context.Context, sessionID string) error {
	// TODO : セッションを取得
	// sess ,err := repo.GetByID(sessionID)

	s.tm.StopTimer(sessionID)

	// TODO : セッションのステートを書き換え
	// err := sess.MoveToPause()
	// err := repo.Store(sess)
	return nil
}

// CanConnectToPusher はイベントをクライアントにプッシュするためのコネクションを貼れるかどうかチェックします。
func (s *SessionUseCase) CanConnectToPusher(sessionID string) (bool, error) {
	// TODO : セッションを取得
	// sess ,err := repo.GetByID(sessionID)

	// TODO 諸々のチェックをする

	// TODO : セッションが再生中なのに同期チェックがされていなかったら始める
	// サーバ再起動でタイマーがなくなると、イベントが正しくクライアントに送られなくなるのでこのタイミングで復旧させる。
	// if _, ok := s.tm.GetTimer(sessionID); !ok && sess.IsPlaying() {
	// 	go s.startSyncCheck(sessionID)
	// }

	return true, nil
}

// TODO 関数名を変える。synccheckだけじゃなくて曲が終わった後のビジネスロジックを賄っているので
// startSyncCheck はSpotifyとの同期が取れているかチェックを行います。goroutineで実行されることを想定しています。
func (s *SessionUseCase) startSyncCheck(ctx context.Context, sessionID string) {
	currentPlayingInfo, err := s.playerCli.CurrentlyPlaying(ctx)
	if err != nil {
		fmt.Printf("startSyncCheck : call currently playing api: %v \n", err)
		return
	}
	triggerAfterTrackEnd := s.tm.CreateTimer(sessionID, currentPlayingInfo.Remain()+syncCheckOffset)

	for {
		select {
		case <-triggerAfterTrackEnd.StopCh():
			fmt.Printf("timer stopped sessionID=%s\n", sessionID)
			return
		case <-triggerAfterTrackEnd.ExpireCh():
			triggerAfterTrackEnd, err = s.syncCheck(ctx, sessionID)
			if err != nil {
				fmt.Printf("synccheck: %v\n", err)
				return
			}
		}
	}
}

func (s *SessionUseCase) syncCheck(ctx context.Context, sessionID string) (*entity.SyncCheckTimer, error) {
	// TODO DBに保存されているセッション情報を取得してキューが全て消化されたらDBのstate_typeをSTOPに変更して戻る
	// sess ,err := repo.GetByID(sessionID)

	// TODO APIトークンの更新
	currentPlayingInfo, err := s.playerCli.CurrentlyPlaying(ctx)
	if err != nil {
		return nil, fmt.Errorf("call currently playing api: %w", err)
	}

	// TODO ずれているかチェック

	triggerAfterTrackEnd := s.tm.CreateTimer(sessionID, currentPlayingInfo.Remain()+syncCheckOffset)
	s.pusher.Push(&event.PushMessage{
		SessionID: sessionID,
		Msg:       entity.EventNextTrack,
	})
	// TODO headを前にすすめる

	{
		s.pusher.Push(&event.PushMessage{
			SessionID: sessionID,
			Msg:       entity.EventInterrupt,
		})
		s.tm.StopTimer(sessionID)
		// TODO : DBのstate_typeを更新
	}
	return triggerAfterTrackEnd, nil
}

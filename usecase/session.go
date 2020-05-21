package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
	"github.com/camphor-/relaym-server/domain/repository"
	"github.com/camphor-/relaym-server/domain/spotify"
)

var syncCheckOffset = 5 * time.Second

// SessionUseCase はセッションに関するユースケースです。
type SessionUseCase struct {
	tm          *entity.SyncCheckTimerManager
	sessionRepo repository.Session
	userRepo    repository.User
	playerCli   spotify.Player
	pusher      event.Pusher
}

// NewSessionUseCase はSessionUseCaseのポインタを生成します。
func NewSessionUseCase(sessionRepo repository.Session, userRepo repository.User, playerCli spotify.Player, pusher event.Pusher) *SessionUseCase {
	return &SessionUseCase{
		tm:          entity.NewSyncCheckTimerManager(),
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		playerCli:   playerCli,
		pusher:      pusher,
	}
}

// CreateSession は与えられたセッション名のセッションを作成します。
func (s *SessionUseCase) CreateSession(sessionName string, creatorID string) (*entity.SessionWithUser, error) {
	creator, err := s.userRepo.FindByID(creatorID)
	if err != nil {
		return nil, fmt.Errorf("CreateSession: %w", err)
	}

	newSession, errNewSession := entity.NewSession(sessionName, creatorID)
	if errNewSession != nil {
		return nil, fmt.Errorf("CreateSession: %w", errNewSession)
	}

	errWithStore := s.sessionRepo.StoreSession(newSession)
	if errWithStore != nil {
		return nil, fmt.Errorf("createSession sessionName=%s: %w", sessionName, errWithStore)
	}
	return entity.SessionToSessionWithUser(newSession, creator), nil
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
	sess, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return fmt.Errorf("find session id=%s: %w", sessionID, err)
	}

	// TODO : デバイスIDをどっかから読み込む
	if err := s.playerCli.Play(ctx, ""); err != nil {
		return fmt.Errorf("call play api: %w", err)
	}

	if err := sess.MoveToPlay(); err != nil {
		return fmt.Errorf("move to play id=%s: %w", sessionID, err)
	}

	if err := s.sessionRepo.Update(sess); err != nil {
		return fmt.Errorf("update session id=%s: %w", sessionID, err)
	}

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
	// TODO : Spotify APIで現在の再生状況を取得
	remainDuration := 3 * time.Minute

	triggerAfterTrackEnd := s.tm.CreateTimer(sessionID, remainDuration+syncCheckOffset)

	for {
		select {
		case <-triggerAfterTrackEnd.StopCh():
			fmt.Printf("timer stopped sessionID=%s\n", sessionID)
			return
		case <-triggerAfterTrackEnd.ExpireCh():
			// TODO DBに保存されているセッション情報を取得
			// TODO : Spotify APIで現在の再生状況を取得
			// 問題なければ新しいtimerをセット
			{
				newD := 4 * time.Minute
				triggerAfterTrackEnd = s.tm.CreateTimer(sessionID, newD+syncCheckOffset)
				s.pusher.Push(&event.PushMessage{
					SessionID: sessionID,
					Msg: &entity.Event{
						Type: "NEXTTRACK",
					},
				})
			}

			// TODO 同期に失敗したらエラーを通知して終了
			{
				s.pusher.Push(&event.PushMessage{
					SessionID: sessionID,
					Msg: &entity.Event{
						Type: "INTERRUPT",
					},
				})
				s.tm.StopTimer(sessionID)
				return
			}

		}
	}
}

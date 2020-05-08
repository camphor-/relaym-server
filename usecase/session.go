package usecase

import (
	"fmt"
	"time"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/event"
)

var syncCheckOffset = 5 * time.Second

// SessionUseCase はセッションに関するユースケースです。
type SessionUseCase struct {
	tm     *entity.SyncCheckTimerManager
	pusher event.Pusher
}

// NewSessionUseCase はSessionUseCaseのポインタを生成します。
func NewSessionUseCase() *SessionUseCase {
	return &SessionUseCase{}
}

// Play はセッションのstateを STOP → PLAY に変更して曲の再生を始めます。
func (s *SessionUseCase) Play(sessionID string) error {
	// TODO : セッションを取得
	// sess ,err := repo.GetByID(sessionID)

	// TODO : Spotify APIで再生
	// playCli.Play(sess.HeadSession())

	// TODO : セッションのステートを書き換え
	// err := sess.MoveToPlay()
	// err := repo.Store(sess)

	go s.startSyncCheck(sessionID)

	return nil
}

// Pause はセッションのstateをPLAY→PAUSEに変更して曲の再生を一時停止します。
func (s *SessionUseCase) Pause(sessionID string) {
	// TODO : セッションを取得
	// sess ,err := repo.GetByID(sessionID)

	s.tm.StopTimer(sessionID)

	// TODO : セッションのステートを書き換え
	// err := sess.MoveToPause()
	// err := repo.Store(sess)

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

// startSyncCheck はSpotifyとの同期が取れているかチェックを行います。goroutineで実行されることを想定しています。
func (s *SessionUseCase) startSyncCheck(sessionID string) {
	// TODO : Spotify APIで現在の再生状況を取得
	remainDuration := 3 * time.Minute

	triggerAfterTrackEnd := s.tm.CreateTimer(sessionID, remainDuration+syncCheckOffset)

	for {
		select {
		case <-triggerAfterTrackEnd.StopCh():
			fmt.Printf("timer stopped sessionID=%s\n", sessionID)
			return
		case <-triggerAfterTrackEnd.ExpireCh():
			// TODO DBに保存されているセッション情報を種痘
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

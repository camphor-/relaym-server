package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/camphor-/relaym-server/domain/entity"
	"github.com/camphor-/relaym-server/domain/mock_spotify"
	"github.com/camphor-/relaym-server/usecase"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/labstack/echo/v4"
)

func TestTrackHandler_SearchTracks(t *testing.T) {
	albumImageJSONs := []*albumImageJSON{
		{
			URL:    "https://i.scdn.co/image/ab67616d0000b273b48630d6efcebca2596120c4",
			Height: 640,
			Width:  640,
		},
	}

	artistJSONs := []*artistJSON{
		{
			Name: "MONOEYES",
		},
	}

	trackJSONs := []*trackJSON{
		{
			URI:      "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
			ID:       "06QTSGUEgcmKwiEJ0IMPig",
			Name:     "Borderland",
			Duration: 213066,
			Artists:  artistJSONs,
			URL:      "https://open.spotify.com/track/06QTSGUEgcmKwiEJ0IMPig",
			Album: &albumJSON{
				Name:   "Interstate 46 E.P.",
				Images: albumImageJSONs,
			},
		},
	}

	tests := []struct {
		name                  string
		prepareQueryFn        func() url.Values
		prepareMockTrackSpoFn func(mock *mock_spotify.MockTrackClient)
		want                  *tracksRes
		wantErr               bool
		wantCode              int
	}{
		{
			name: "qが存在するとき正しく動作する",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("q", "MONOEYES")
				return q
			},
			prepareMockTrackSpoFn: func(mock *mock_spotify.MockTrackClient) {
				artists := []*entity.Artist{
					{
						Name: "MONOEYES",
					},
				}

				albumImages := []*entity.AlbumImage{
					{
						URL:    "https://i.scdn.co/image/ab67616d0000b273b48630d6efcebca2596120c4",
						Height: 640,
						Width:  640,
					},
				}

				tracks := []*entity.Track{
					{
						URI:      "spotify:track:06QTSGUEgcmKwiEJ0IMPig",
						ID:       "06QTSGUEgcmKwiEJ0IMPig",
						Name:     "Borderland",
						Duration: 213066000000,
						Artists:  artists,
						URL:      "https://open.spotify.com/track/06QTSGUEgcmKwiEJ0IMPig",
						Album: &entity.Album{
							Name:   "Interstate 46 E.P.",
							Images: albumImages,
						},
					},
				}
				mock.EXPECT().Search(gomock.Any(), "MONOEYES").Return(tracks, nil)
			},
			want: &tracksRes{
				Tracks: trackJSONs,
			},
			wantErr:  false,
			wantCode: http.StatusOK,
		},
		{
			name: "qが存在しないときStatusCodeで400が返る",
			prepareQueryFn: func() url.Values {
				q := url.Values{}
				q.Set("q", "")
				return q
			},
			prepareMockTrackSpoFn: func(mock *mock_spotify.MockTrackClient) {},
			want:                  nil,
			wantErr:               true,
			wantCode:              http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptestの準備
			e := echo.New()
			q := tt.prepareQueryFn()
			req := httptest.NewRequest(http.MethodPost, "/?"+q.Encode(), nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// モックの準備
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockTrackSpo := mock_spotify.NewMockTrackClient(ctrl)
			tt.prepareMockTrackSpoFn(mockTrackSpo)
			uc := usecase.NewTrackUseCase(mockTrackSpo)
			h := &TrackHandler{trackUC: uc}

			err := h.SearchTracks(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("TrackHandler.SearchTracks() error = %v, wantErr %v", err, tt.wantErr)
			}

			// ステータスコードのチェック
			if er, ok := err.(*echo.HTTPError); (ok && er.Code != tt.wantCode) || (!ok && rec.Code != tt.wantCode) {
				t.Errorf("TrackHandler.SearchTracks() code = %d, want = %d", rec.Code, tt.wantCode)
			}
			if !tt.wantErr {
				got := &tracksRes{}
				err := json.Unmarshal(rec.Body.Bytes(), got)
				if err != nil {
					t.Fatal(err)
				}
				if !cmp.Equal(got, tt.want) {
					t.Errorf("TrackHandler.SearchTracks() diff = %v", cmp.Diff(got, tt.want))
				}
			}
		})
	}
}

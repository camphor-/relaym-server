package entity

import "time"

// Track は曲を表す構造体です。
type Track struct {
	URI      string
	ID       string
	Name     string
	Duration time.Duration
	Artists  []*Artist
	URL      string // Spotifyのwebページ
	Album    *Album
}

type Album struct {
	Name   string
	Images []*AlbumImage
}

type AlbumImage struct {
	URL    string
	Height int
	Width  int
}

type Artist struct {
	Name string
}

// CurrentPlayingInfo は現在再生している情報を表します
type CurrentPlayingInfo struct {
	Playing  bool
	Progress time.Duration
	Track    *Track
}

// Remain は残りの再生時間を計算します。
func (cpi *CurrentPlayingInfo) Remain() time.Duration {
	return cpi.Track.Duration - cpi.Progress
}

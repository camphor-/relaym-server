package entity

import "time"

// Trackは曲を表す構造体です。
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

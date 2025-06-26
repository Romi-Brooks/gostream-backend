package models

type MusicFile struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	FileType string `json:"fileType"`
	Size     int64  `json:"size"`
}

type MusicMetadata struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	Year     int    `json:"year"`
	Genre    string `json:"genre"`
	AlbumArt string `json:"albumArt"`
}

// 更新歌词响应结构
type LyricResponse struct {
	Lyrics string `json:"lyrics"`
	Source string `json:"source"` // local, cached, online, none
}

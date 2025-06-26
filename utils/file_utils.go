package utils

import (
	"backend/models"
	"os"
	"path/filepath"
	"strings"

	"bytes"
	"encoding/base64"
	"github.com/dhowden/tag"
	"github.com/disintegration/imaging"
	"image"
	"image/jpeg"
)

func GetMusicFiles(musicDir string) ([]models.MusicFile, error) {
	var musicFiles []models.MusicFile

	err := filepath.Walk(musicDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && IsMusicFile(path) {
			relPath, _ := filepath.Rel(musicDir, path)

			musicFiles = append(musicFiles, models.MusicFile{
				Name:     info.Name(),
				Path:     relPath,
				FileType: strings.TrimPrefix(filepath.Ext(path), "."),
				Size:     info.Size(),
			})
		}
		return nil
	})

	return musicFiles, err
}

func IsMusicFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp3", ".flac", ".wav", ".ogg", ".m4a":
		return true
	}
	return false
}

func GetMusicMetadata(filePath string) (models.MusicMetadata, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return models.MusicMetadata{}, err
	}
	defer file.Close()

	// 解析元数据
	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return models.MusicMetadata{}, err
	}

	// 创建响应对象
	response := models.MusicMetadata{
		Title:  metadata.Title(),
		Artist: metadata.Artist(),
		Album:  metadata.Album(),
		Year:   metadata.Year(),
		Genre:  metadata.Genre(),
	}

	// 处理专辑封面
	if pic := metadata.Picture(); pic != nil {
		img, _, err := image.Decode(bytes.NewReader(pic.Data))
		if err == nil {
			// 调整图片大小 (300x300)
			resizedImg := imaging.Resize(img, 300, 300, imaging.Lanczos)

			// 编码为JPEG
			var buf bytes.Buffer
			if err := jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: 85}); err == nil {
				response.AlbumArt = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
			}
		}
	}

	return response, nil
}

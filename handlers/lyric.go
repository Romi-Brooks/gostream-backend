package handlers

import (
	"backend/config"
	"backend/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type LyricsSearchResult struct {
	ID           json.Number `json:"id"` // 可以处理数字和字符串
	TrackName    string      `json:"trackName"`
	ArtistName   string      `json:"artistName"`
	AlbumName    string      `json:"albumName"`
	Duration     float64     `json:"duration"`
	Instrumental bool        `json:"instrumental"`
	PlainLyrics  string      `json:"plainLyrics"`
	SyncedLyrics string      `json:"syncedLyrics"`
}

type LyricsGetResponse struct {
	ID           json.Number `json:"id"`
	SyncedLyrics string      `json:"syncedLyrics"`
}

// 添加缓存结构
var lyricsCache = make(map[string]string) // key: trackID, value: syncedLyrics

// 本地歌词获取函数
func getLocalLyrics(filePath string) (string, error) {
	fullPath := filepath.Join(config.MusicDir, filePath)
	lyricPath := strings.TrimSuffix(fullPath, filepath.Ext(fullPath)) + ".lrc"

	// 检查歌词文件是否存在
	if _, err := os.Stat(lyricPath); os.IsNotExist(err) {
		return "", fmt.Errorf("lyric file not found")
	}

	// 读取歌词文件
	lyricBytes, err := ioutil.ReadFile(lyricPath)
	if err != nil {
		return "", err
	}

	return string(lyricBytes), nil
}

func GetLyricHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 获取歌曲信息
	title := r.URL.Query().Get("title")
	artist := r.URL.Query().Get("artist")
	filePath := r.URL.Query().Get("path")

	if filePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}

	// 尝试获取本地歌词
	localLyrics, err := getLocalLyrics(filePath)
	if err == nil {
		response := models.LyricResponse{
			Lyrics: localLyrics,
			Source: "local",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Local lyrics not found: %v", err)

	// 如果本地没有，尝试从网络获取
	if title == "" || artist == "" {
		http.Error(w, "Title and artist required for online search", http.StatusBadRequest)
		return
	}

	onlineLyrics, source, err := getOnlineLyrics(title, artist, filePath)
	if err != nil {
		log.Printf("Error fetching online lyrics: %v", err)
		// 返回空歌词，由前端提示用户上传
		response := models.LyricResponse{
			Lyrics: "",
			Source: "none",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.LyricResponse{
		Lyrics: onlineLyrics,
		Source: source,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 更新在线歌词获取逻辑
func getOnlineLyrics(title, artist, filePath string) (string, string, error) {
	searchResults, err := searchLyrics(title, artist)
	if err != nil {
		return "", "online", err
	}

	if len(searchResults) == 0 {
		return "", "online", fmt.Errorf("no lyrics found for %s - %s", artist, title)
	}

	topResult := searchResults[0]
	idStr := topResult.ID.String()

	if cachedLyrics, ok := lyricsCache[idStr]; ok {
		return cachedLyrics, "cached", nil
	}

	syncedLyrics, err := getSyncedLyrics(idStr)
	if err != nil {
		return "", "online", err
	}

	lyricsCache[idStr] = syncedLyrics
	saveLyricsToLocal(filePath, syncedLyrics)

	return syncedLyrics, "online", nil
}

// 搜索歌词
func searchLyrics(title, artist string) ([]LyricsSearchResult, error) {
	// 构建查询参数
	params := url.Values{}

	// 优先使用精确匹配
	if title != "" && artist != "" {
		params.Add("track_name", title)
		params.Add("artist_name", artist)
	} else if title != "" {
		params.Add("q", title)
	} else if artist != "" {
		params.Add("q", artist)
	} else {
		return nil, fmt.Errorf("insufficient search parameters")
	}

	// 构建API URL
	apiURL := fmt.Sprintf("%s/api/search?%s", config.LyricsAPIBaseURL, params.Encode())

	// 发送请求
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	// 解析响应
	var results []LyricsSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	return results, nil
}

func getSyncedLyrics(id string) (string, error) {
	apiURL := fmt.Sprintf("%s/api/get/%s", config.LyricsAPIBaseURL, id)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status: %d", resp.StatusCode)
	}

	var response LyricsGetResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.SyncedLyrics, nil
}

// 保存歌词到本地文件
func saveLyricsToLocal(filePath, lyrics string) {
	fullPath := filepath.Join(config.MusicDir, filePath)
	lyricPath := strings.TrimSuffix(fullPath, filepath.Ext(fullPath)) + ".lrc"

	// 创建目录（如果不存在）
	dir := filepath.Dir(lyricPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	// 写入文件
	err := ioutil.WriteFile(lyricPath, []byte(lyrics), 0644)
	if err != nil {
		log.Printf("Failed to save lyrics to local: %v", err)
	}
}

func UploadLyricHandler(w http.ResponseWriter, r *http.Request) {
	EnableCors(&w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 解析表单
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	filePath := r.FormValue("path")
	if filePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}

	// 获取上传的文件
	file, _, err := r.FormFile("lyric")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 读取文件内容
	lyricBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// 保存歌词文件
	fullPath := filepath.Join(config.MusicDir, filePath)
	lyricPath := strings.TrimSuffix(fullPath, filepath.Ext(fullPath)) + ".lrc"

	err = ioutil.WriteFile(lyricPath, lyricBytes, 0644)
	if err != nil {
		log.Printf("Error saving lyric file: %v", err)
		http.Error(w, "Error saving lyric file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Lyrics uploaded successfully"))
}

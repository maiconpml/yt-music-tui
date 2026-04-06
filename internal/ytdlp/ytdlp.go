package ytdlp

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

var (
	activeCookiePath string
	activeCacheDir   string
	cacheQueue       []string
	cacheMutex       sync.Mutex
	maxCacheSize     = 10
)

func registerCache(videoID string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	for i, id := range cacheQueue {
		if id == videoID {
			cacheQueue = append(cacheQueue[:i], cacheQueue[i+1:]...)
			break
		}
	}

	cacheQueue = append(cacheQueue, videoID)

	for len(cacheQueue) > maxCacheSize {
		oldest := cacheQueue[0]
		cacheQueue = cacheQueue[1:]
		err := os.Remove(filepath.Join(activeCacheDir, oldest+".mp3"))
		if err != nil && !os.IsNotExist(err) {
			slog.Error("Erro ao remover cache antigo", "videoID", oldest, "err", err)
		} else {
			slog.Info("Cache antigo removido", "videoID", oldest)
		}
	}
}

func CleanCache() {
	if activeCacheDir != "" {
		os.RemoveAll(activeCacheDir)
	}
}

func Init(cookiePath string) error {
	activeCookiePath = cookiePath

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}

	activeCacheDir = filepath.Join(cacheDir, "heylisten")
	if err := os.MkdirAll(activeCacheDir, 0o755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}
	return nil
}

// GetAudioPath downloads the audio for the given videoID to a cache directory
// and returns the path to the downloaded MP3 file.
func GetAudioPath(videoID string) (string, error) {
	path := filepath.Join(activeCacheDir, videoID+".mp3")
	if _, err := os.Stat(path); err == nil {
		registerCache(videoID)
		return path, nil
	}

	url := fmt.Sprintf("https://music.youtube.com/watch?v=%s", videoID)
	// -x: extract audio, --audio-format mp3: convert to mp3, --audio-quality 0: best quality
	args := []string{"-x", "--audio-format", "mp3", "--audio-quality", "0", "--output", path, "--postprocessor-args", "ffmpeg:-ar 44100"}
	if activeCookiePath != "" {
		args = append(args, "--cookies", activeCookiePath)
	}
	args = append(args, url)

	cmd := exec.Command("yt-dlp", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("yt-dlp failed: %v, output: %s", err, string(output))
	}

	slog.Info("Download concluído (cache)", "videoID", videoID)
	registerCache(videoID)
	return path, nil
}

func StreamAudio(videoID string) (io.Reader, error) {
	path := filepath.Join(activeCacheDir, videoID+".mp3")
	if _, err := os.Stat(path); err == nil {
		slog.Info("Usando arquivo de cache existente", "videoID", videoID)
		registerCache(videoID)
		return os.Open(path)
	}

	url := fmt.Sprintf("https://music.youtube.com/watch?v=%s", videoID)

	ytCmd := exec.Command("yt-dlp", "-q", "--no-progress", "-o", "-", "-f", "ba", url)
	if activeCookiePath != "" {
		ytCmd.Args = append(ytCmd.Args, "--cookies", activeCookiePath)
	}

	ffCmd := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "mp3", "-ar", "44100", "-map_metadata", "-1", "-id3v2_version", "0", "-")

	ytOut, err := ytCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	ffCmd.Stdin = ytOut

	var ffStderr bytes.Buffer
	ffCmd.Stderr = &ffStderr
	ffOut, err := ffCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	if err := ytCmd.Start(); err != nil {
		file.Close()
		return nil, err
	}
	if err := ffCmd.Start(); err != nil {
		file.Close()
		return nil, err
	}

	registerCache(videoID)

	done := make(chan struct{})
	go func() {
		defer file.Close()
		defer close(done)

		_, copyErr := io.Copy(file, ffOut)

		_ = ytCmd.Wait()
		_ = ffCmd.Wait()

		if copyErr == nil {
			slog.Info("Download e conversão concluídos (cache pronto)", "videoID", videoID)
		} else {
			slog.Error("Erro durante cache assíncrono", "error", copyErr)
		}
	}()

	readHandle, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return &followReader{
		file: readHandle,
		done: done,
	}, nil
}

type followReader struct {
	file *os.File
	done chan struct{}
}

func (f *followReader) Read(p []byte) (n int, err error) {
	for {
		n, err = f.file.Read(p)
		if n > 0 {
			return n, nil
		}
		if err == io.EOF {
			select {
			case <-f.done:
				return 0, io.EOF
			default:
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
		return n, err
	}
}

func (f *followReader) Close() error {
	return f.file.Close()
}

// CheckDependencies checks if yt-dlp and ffmpeg are installed.
func CheckDependencies() error {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return fmt.Errorf("yt-dlp is not installed. Please install it to download audio")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg is not installed. It is required by yt-dlp for audio extraction")
	}
	return nil
}

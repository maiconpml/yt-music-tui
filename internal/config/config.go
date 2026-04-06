package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(configDir, "heylisten")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func GetCookiePath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "cookie.txt"), nil
}

func GetLogPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "app.log"), nil
}

func GetAuthUserPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "authuser.txt"), nil
}

func SaveAuthUser(authUser string) error {
	path, err := GetAuthUserPath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.TrimSpace(authUser)), 0o644)
}

func LoadAuthUser() (string, error) {
	path, err := GetAuthUserPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func SaveCookie(rawCookie string) error {
	path, err := GetCookiePath()
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("# Netscape HTTP Cookie File\n")
	f.WriteString("# This is a generated file! Do not edit.\n\n")

	parts := strings.Split(rawCookie, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		name := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		if name != "" && value != "" {
			// Format: domain  include_subdomains  path  is_secure  expiry  name  value
			// Using 0 for expiry (session) and TRUE for secure
			line := fmt.Sprintf(".youtube.com\tTRUE\t/\tTRUE\t0\t%s\t%s\n", name, value)
			f.WriteString(line)
			lineMusic := fmt.Sprintf(".music.youtube.com\tTRUE\t/\tTRUE\t0\t%s\t%s\n", name, value)
			f.WriteString(lineMusic)
		}
	}
	return nil
}

func LoadCookie() (string, error) {
	path, err := GetCookiePath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	content := string(data)
	if strings.HasPrefix(content, "# Netscape") || strings.Contains(content, "\t") {
		return parseNetscapeCookies(content), nil
	}

	return strings.TrimSpace(content), nil
}

func parseNetscapeCookies(content string) string {
	cookieMap := make(map[string]string)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 7 {
			name := strings.TrimSpace(parts[5])
			value := strings.TrimSpace(parts[6])
			if name != "" && value != "" {
				// Deduplicate by name, keeping the first or last found
				cookieMap[name] = value
			}
		}
	}

	var cookies []string
	for name, value := range cookieMap {
		cookies = append(cookies, fmt.Sprintf("%s=%s", name, value))
	}
	return strings.Join(cookies, "; ")
}

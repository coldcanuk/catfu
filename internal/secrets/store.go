// Package secrets stores credentials in the OS keychain when available,
// with a restricted local file fallback for headless environments.
package secrets

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	serviceName     = "catfu"
	braveKeyAccount = "brave_api_key"
)

// Source describes where a secret was resolved from.
type Source string

const (
	SourceNone    Source = ""
	SourceFlag    Source = "flag"
	SourceEnv     Source = "env"
	SourceKeyring Source = "keyring"
	SourceFile    Source = "file"
	SourceConfig  Source = "config"
)

// ErrNotFound is returned when a secret is missing.
var ErrNotFound = errors.New("secret not found")

// Backend reports which storage backend is active for writes.
type Backend string

const (
	BackendKeyring Backend = "keyring"
	BackendFile    Backend = "file"
)

// SetBraveAPIKey stores the Brave Search plan token.
// Prefers OS keyring; falls back to a 0600 file under the config dir.
func SetBraveAPIKey(token string) (Backend, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("token is empty")
	}
	if err := keyring.Set(serviceName, braveKeyAccount, token); err == nil {
		// Best-effort: remove file copy so keyring is sole source.
		_ = os.Remove(filePath())
		return BackendKeyring, nil
	}
	if err := writeFileSecret(token); err != nil {
		return "", fmt.Errorf("keyring unavailable and file store failed: %w", err)
	}
	return BackendFile, nil
}

// GetBraveAPIKey reads the stored token. Returns ErrNotFound if absent.
func GetBraveAPIKey() (string, Source, error) {
	if v, err := keyring.Get(serviceName, braveKeyAccount); err == nil && strings.TrimSpace(v) != "" {
		return v, SourceKeyring, nil
	} else if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		// keyring backend error — try file
	}
	v, err := readFileSecret()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", SourceNone, ErrNotFound
		}
		return "", SourceNone, err
	}
	if v == "" {
		return "", SourceNone, ErrNotFound
	}
	return v, SourceFile, nil
}

// DeleteBraveAPIKey removes the token from keyring and file store.
func DeleteBraveAPIKey() error {
	var first error
	if err := keyring.Delete(serviceName, braveKeyAccount); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		// ignore missing; note other errors but still try file
		first = err
	}
	if err := os.Remove(filePath()); err != nil && !os.IsNotExist(err) {
		if first == nil {
			first = err
		}
	}
	return first
}

// Status reports presence without returning the secret.
func Status() map[string]any {
	out := map[string]any{
		"service": serviceName,
		"account": braveKeyAccount,
	}
	if _, src, err := GetBraveAPIKey(); err == nil {
		out["brave_api_key_set"] = true
		out["brave_api_key_source"] = string(src)
	} else {
		out["brave_api_key_set"] = false
		out["brave_api_key_source"] = ""
	}
	// Probe keyring availability
	if err := keyring.Set(serviceName, "__catfu_probe__", "1"); err == nil {
		_ = keyring.Delete(serviceName, "__catfu_probe__")
		out["keyring_available"] = true
		out["preferred_backend"] = string(BackendKeyring)
	} else {
		out["keyring_available"] = false
		out["keyring_error"] = err.Error()
		out["preferred_backend"] = string(BackendFile)
		out["file_path"] = filePath()
	}
	return out
}

func filePath() string {
	dir := configDir()
	return filepath.Join(dir, "secrets")
}

func configDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "catfu")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "catfu"
	}
	return filepath.Join(home, ".config", "catfu")
}

func writeFileSecret(token string) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	// Simple KEY=value format; only brave_api_key for now.
	content := "brave_api_key=" + token + "\n"
	path := filePath()
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readFileSecret() (string, error) {
	b, err := os.ReadFile(filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotFound
		}
		return "", err
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok && strings.TrimSpace(k) == "brave_api_key" {
			return strings.TrimSpace(v), nil
		}
	}
	return "", ErrNotFound
}

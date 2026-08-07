package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	// Force file path under temp by clearing any real keyring use for this test:
	// writeFileSecret uses configDir which reads XDG_CONFIG_HOME.
	if err := writeFileSecret("tok-abc"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "catfu", "secrets")
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm()&0o077 != 0 {
		t.Fatalf("secrets file too permissive: %v", fi.Mode())
	}
	v, err := readFileSecret()
	if err != nil || v != "tok-abc" {
		t.Fatalf("got %q err=%v", v, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	_, err = readFileSecret()
	if err != ErrNotFound {
		t.Fatalf("want ErrNotFound got %v", err)
	}
}

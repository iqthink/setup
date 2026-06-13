package profile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func withTempHome(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
}

func TestExistsAndLoad_MissingFile(t *testing.T) {
	withTempHome(t)
	if Exists() {
		t.Fatal("Exists() returned true for an empty home dir")
	}
	picks, err := Load()
	if err != nil {
		t.Fatalf("Load on missing file errored: %v", err)
	}
	if picks != nil {
		t.Fatalf("Load on missing file returned %v, want nil", picks)
	}
}

func TestSaveAndLoad_Roundtrip(t *testing.T) {
	withTempHome(t)
	want := []string{"raycast", "figma", "zed"}
	if err := Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if !Exists() {
		t.Fatal("Exists() false after Save")
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load returned %v, want %v", got, want)
	}
}

func TestSave_AtomicReplace(t *testing.T) {
	withTempHome(t)
	if err := Save([]string{"raycast"}); err != nil {
		t.Fatal(err)
	}
	if err := Save([]string{"figma", "zed"}); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"figma", "zed"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	// No stray .tmp file should remain after a successful Save.
	p, _ := Path()
	if _, err := os.Stat(p + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("expected no .tmp residue, got err=%v", err)
	}
}

func TestSave_EmptyPicksStillCreatesFile(t *testing.T) {
	withTempHome(t)
	if err := Save(nil); err != nil {
		t.Fatal(err)
	}
	if !Exists() {
		t.Fatal("Exists() false after Save(nil); first-run skip-everything should still mark the user as picked")
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("Load after Save(nil) returned %v, want empty", got)
	}
}

func TestLoad_IgnoresCommentsAndBlankLines(t *testing.T) {
	withTempHome(t)
	p, _ := Path()
	body := "# header\n\nraycast\n  # indented comment with whitespace\n\nfigma\n"
	if err := os.WriteFile(filepath.Join(p), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"raycast", "figma"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

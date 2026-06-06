package main

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestHome points HOME at a fresh temp dir for the duration of the test,
// then restores it automatically via t.Cleanup.
func setupTestHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	initData()
	return dir
}

// --- pure functions ---

func TestSplitKV(t *testing.T) {
	cases := []struct {
		input    string
		wantKey  string
		wantVal  string
		wantOK   bool
	}{
		{"pin0=/home/user/file.txt", "pin0", "/home/user/file.txt", true},
		{"pager=bat", "pager", "bat", true},
		// value containing '=' should not be split further
		{"pin1=/path/with=equals", "pin1", "/path/with=equals", true},
		// no '=' → not a valid pair
		{"notavalidline", "", "", false},
		{"", "", "", false},
	}
	for _, c := range cases {
		k, v, ok := splitKV(c.input)
		if ok != c.wantOK || k != c.wantKey || v != c.wantVal {
			t.Errorf("splitKV(%q) = (%q, %q, %v), want (%q, %q, %v)",
				c.input, k, v, ok, c.wantKey, c.wantVal, c.wantOK)
		}
	}
}

func TestValidatePager(t *testing.T) {
	valid := []string{"bat", "less", "more", "most", "cat", "pg"}
	for _, p := range valid {
		if !validatePager(p) {
			t.Errorf("validatePager(%q) = false, want true", p)
		}
	}
	invalid := []string{"vim", "nano", "emacs", "", "b at", "BAT"}
	for _, p := range invalid {
		if validatePager(p) {
			t.Errorf("validatePager(%q) = true, want false", p)
		}
	}
}

func TestIsDigit(t *testing.T) {
	for c := byte('0'); c <= '9'; c++ {
		if !isDigit(c) {
			t.Errorf("isDigit(%q) = false, want true", c)
		}
	}
	for _, c := range []byte{'a', 'z', '/', '-', ' '} {
		if isDigit(c) {
			t.Errorf("isDigit(%q) = true, want false", c)
		}
	}
}

func TestSlotKey(t *testing.T) {
	cases := map[byte]string{
		'0': "pin0",
		'5': "pin5",
		'9': "pin9",
	}
	for slot, want := range cases {
		if got := slotKey(slot); got != want {
			t.Errorf("slotKey(%q) = %q, want %q", slot, got, want)
		}
	}
}

// --- data layer ---

func TestInitData(t *testing.T) {
	home := setupTestHome(t)
	dir := filepath.Join(home, dataSubdir)
	data := filepath.Join(dir, dataFile)

	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("data directory not created: %v", err)
	}
	if _, err := os.Stat(data); err != nil {
		t.Fatalf("data file not created: %v", err)
	}

	// calling initData again must not error or duplicate the file
	initData()
	if _, err := os.Stat(data); err != nil {
		t.Fatalf("data file missing after second initData: %v", err)
	}
}

func TestReadKeyMissing(t *testing.T) {
	setupTestHome(t)
	if got := readKey("pin0"); got != "" {
		t.Errorf("readKey on empty store = %q, want empty string", got)
	}
}

func TestWriteAndReadKey(t *testing.T) {
	setupTestHome(t)
	writeKey("pin0", "/tmp/foo.txt")
	if got := readKey("pin0"); got != "/tmp/foo.txt" {
		t.Errorf("readKey = %q, want %q", got, "/tmp/foo.txt")
	}
}

func TestWriteKeyOverwrites(t *testing.T) {
	setupTestHome(t)
	writeKey("pin0", "/tmp/first.txt")
	writeKey("pin0", "/tmp/second.txt")
	if got := readKey("pin0"); got != "/tmp/second.txt" {
		t.Errorf("after overwrite readKey = %q, want %q", got, "/tmp/second.txt")
	}
}

func TestMultipleKeysCoexist(t *testing.T) {
	setupTestHome(t)
	writeKey("pin0", "/tmp/a.txt")
	writeKey("pin1", "/tmp/b.txt")
	writeKey("pager", "bat")

	if got := readKey("pin0"); got != "/tmp/a.txt" {
		t.Errorf("pin0 = %q, want /tmp/a.txt", got)
	}
	if got := readKey("pin1"); got != "/tmp/b.txt" {
		t.Errorf("pin1 = %q, want /tmp/b.txt", got)
	}
	if got := readKey("pager"); got != "bat" {
		t.Errorf("pager = %q, want bat", got)
	}
}

func TestDeleteKey(t *testing.T) {
	setupTestHome(t)
	writeKey("pin0", "/tmp/foo.txt")
	deleteKey("pin0")
	if got := readKey("pin0"); got != "" {
		t.Errorf("after deleteKey, readKey = %q, want empty string", got)
	}
}

func TestDeleteKeyLeavesOthersIntact(t *testing.T) {
	setupTestHome(t)
	writeKey("pin0", "/tmp/a.txt")
	writeKey("pin1", "/tmp/b.txt")
	deleteKey("pin0")

	if got := readKey("pin0"); got != "" {
		t.Errorf("deleted key pin0 = %q, want empty string", got)
	}
	if got := readKey("pin1"); got != "/tmp/b.txt" {
		t.Errorf("unrelated key pin1 = %q, want /tmp/b.txt", got)
	}
}

func TestDeleteKeyMissingIsNoop(t *testing.T) {
	setupTestHome(t)
	writeKey("pin0", "/tmp/foo.txt")
	deleteKey("pin9") // pin9 was never written
	if got := readKey("pin0"); got != "/tmp/foo.txt" {
		t.Errorf("pin0 after no-op delete = %q, want /tmp/foo.txt", got)
	}
}

func TestPagerPreference(t *testing.T) {
	setupTestHome(t)

	// no preference stored → empty
	if got := readKey("pager"); got != "" {
		t.Errorf("pager before write = %q, want empty", got)
	}

	writeKey("pager", "less")
	if got := readKey("pager"); got != "less" {
		t.Errorf("pager after write = %q, want less", got)
	}

	deleteKey("pager")
	if got := readKey("pager"); got != "" {
		t.Errorf("pager after delete = %q, want empty", got)
	}
}

func TestDetectPagerReturnsSomething(t *testing.T) {
	p := detectPager()
	if p == "" {
		t.Error("detectPager returned empty string")
	}
	// result must always be a whitelisted pager
	if !validatePager(p) {
		t.Errorf("detectPager returned %q which is not a known pager", p)
	}
}

func TestAllSlotsIndependent(t *testing.T) {
	setupTestHome(t)

	paths := make(map[byte]string)
	for i := byte('0'); i <= '9'; i++ {
		p := "/tmp/file" + string(i) + ".txt"
		paths[i] = p
		writeKey(slotKey(i), p)
	}
	for i := byte('0'); i <= '9'; i++ {
		if got := readKey(slotKey(i)); got != paths[i] {
			t.Errorf("slot %c = %q, want %q", i, got, paths[i])
		}
	}
}

func TestClearAllSlots(t *testing.T) {
	setupTestHome(t)

	for i := 0; i < maxSlots; i++ {
		writeKey(slotKey(byte('0'+i)), "/tmp/file.txt")
	}
	for i := 0; i < maxSlots; i++ {
		deleteKey(slotKey(byte('0' + i)))
	}
	for i := 0; i < maxSlots; i++ {
		if got := readKey(slotKey(byte('0' + i))); got != "" {
			t.Errorf("slot %d not cleared, got %q", i, got)
		}
	}
}

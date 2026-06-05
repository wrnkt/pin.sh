package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	dataSubdir  = ".local/share/pin"
	dataFile    = "pins.data"
	maxSlots    = 10
)

var knownPagers = []string{"bat", "less", "more", "most", "cat", "pg"}

func dataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		die("cannot determine home directory: %v", err)
	}
	return filepath.Join(home, dataSubdir)
}

func dataPath() string {
	return filepath.Join(dataDir(), dataFile)
}

func initData() {
	dir := dataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		die("cannot create data directory: %v", err)
	}
	path := dataPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			die("cannot create data file: %v", err)
		}
		f.Close()
	}
}

func readKey(key string) string {
	f, err := os.Open(dataPath())
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if k, v, ok := splitKV(scanner.Text()); ok && k == key {
			return v
		}
	}
	return ""
}

func writeKey(key, value string) {
	path := dataPath()
	var lines []string
	if f, err := os.Open(path); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if k, _, ok := splitKV(line); ok && k == key {
				continue
			}
			if line != "" {
				lines = append(lines, line)
			}
		}
		f.Close()
	}
	lines = append(lines, key+"="+value)
	writeLines(path, lines)
}

func deleteKey(key string) {
	path := dataPath()
	var lines []string
	if f, err := os.Open(path); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if k, _, ok := splitKV(line); ok && k == key {
				continue
			}
			if line != "" {
				lines = append(lines, line)
			}
		}
		f.Close()
	}
	writeLines(path, lines)
}

func writeLines(path string, lines []string) {
	f, err := os.Create(path)
	if err != nil {
		die("cannot write data file: %v", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

func splitKV(line string) (key, value string, ok bool) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func validatePager(name string) bool {
	for _, p := range knownPagers {
		if p == name {
			return true
		}
	}
	return false
}

func detectPager() string {
	for _, p := range []string{"bat", "less", "cat"} {
		if _, err := exec.LookPath(p); err == nil {
			return p
		}
	}
	return "cat"
}

func openFile(path string) {
	pager := readKey("pager")
	if pager == "" {
		pager = detectPager()
	}
	cmd := exec.Command(pager, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		die("pager exited: %v", err)
	}
}

func slotKey(n byte) string {
	return fmt.Sprintf("pin%c", n)
}

func printHelp() {
	fmt.Print(`pin - pin files to numbered slots for quick access

Usage:
  pin -p <file>        pin file to slot 0
  pin -p<N> <file>     pin file to slot N (1-9)
  pin -<N>             open file in slot N
  pin --list           list all pinned files
  pin -c               clear all slots
  pin -c<N>            clear slot N
  pin --pager <name>   set preferred pager (bat, less, more, most, cat, pg)
  pin --pager-clear    reset to auto-detected pager
  pin --uninstall      remove all pin data
  pin -h, --help       show this help

Data stored in ~/.local/share/pin/pins.data
`)
}

func die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "pin: "+format+"\n", args...)
	os.Exit(1)
}

func main() {
	initData()

	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		return
	}

	flag := args[0]

	switch flag {
	case "-h", "--help":
		printHelp()

	case "--list":
		pager := readKey("pager")
		if pager == "" {
			pager = "(auto: " + detectPager() + ")"
		}
		fmt.Printf("pager: %s\n\n", pager)
		for i := 0; i < maxSlots; i++ {
			val := readKey(fmt.Sprintf("pin%d", i))
			if val != "" {
				fmt.Printf("  [%d] %s\n", i, val)
			} else {
				fmt.Printf("  [%d] (empty)\n", i)
			}
		}

	case "--pager":
		if len(args) < 2 {
			die("--pager requires a pager name")
		}
		name := args[1]
		if !validatePager(name) {
			die("unknown pager %q; valid options: %s", name, strings.Join(knownPagers, ", "))
		}
		writeKey("pager", name)
		fmt.Printf("pin: pager set to %s\n", name)

	case "--pager-clear":
		deleteKey("pager")
		fmt.Println("pin: pager preference cleared")

	case "--uninstall":
		fmt.Print("pin: remove all pin data? [y/N] ")
		var answer string
		fmt.Scanln(&answer)
		if answer == "y" || answer == "Y" {
			os.RemoveAll(dataDir())
			fmt.Println("pin: data removed")
		} else {
			fmt.Println("pin: cancelled")
		}

	case "-c":
		for i := 0; i < maxSlots; i++ {
			deleteKey(fmt.Sprintf("pin%d", i))
		}
		fmt.Println("pin: all slots cleared")

	case "-p":
		if len(args) < 2 {
			die("-p requires a file path")
		}
		pinFile('0', args[1])

	default:
		switch {
		// -<N>: open slot N
		case len(flag) == 2 && flag[0] == '-' && isDigit(flag[1]):
			n := flag[1]
			path := readKey(slotKey(n))
			if path == "" {
				die("slot %c is empty", n)
			}
			if _, err := os.Stat(path); err != nil {
				die("file no longer exists: %s", path)
			}
			openFile(path)

		// -c<N>: clear slot N
		case len(flag) == 3 && flag[0] == '-' && flag[1] == 'c' && isDigit(flag[2]):
			n := flag[2]
			deleteKey(slotKey(n))
			fmt.Printf("pin: slot %c cleared\n", n)

		// -p<N> <file>: pin to slot N
		case len(flag) == 3 && flag[0] == '-' && flag[1] == 'p' && isDigit(flag[2]):
			if len(args) < 2 {
				die("%s requires a file path", flag)
			}
			pinFile(flag[2], args[1])

		default:
			die("unknown flag %q — run 'pin --help' for usage", flag)
		}
	}
}

func pinFile(slot byte, target string) {
	abs, err := filepath.Abs(target)
	if err != nil {
		die("cannot resolve path: %v", err)
	}
	if _, err := os.Stat(abs); err != nil {
		die("file not found: %s", target)
	}
	writeKey(slotKey(slot), abs)
	fmt.Printf("pin: pinned %s to slot %c\n", abs, slot)
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

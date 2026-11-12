package provider

import (
	"bytes"
	"os"
	"testing"

	"github.com/yetone/magpie/internal/filememo"
)

// A write of what logins.json already holds changes nothing but its time,
// and filememo forgets what it parsed: every read after it parses the
// accounts again. A gateway was read 230 KB of accounts 182 times in 15
// seconds (strace) after such a write, most of its 20% CPU. A file already
// holding the bytes the write would put there is left alone, and what was
// parsed of it stays.
func TestWriteLoginsLeavesAnUnchangedFileAlone(t *testing.T) {
	signIn(t)
	rememberLogins(true)

	parses := 0
	read := func() {
		_, _ = filememo.Read("logins_unchanged_test", loginsPath(), func(b []byte) (int, error) {
			parses++
			return len(b), nil
		})
	}
	read()

	loginsMu.Lock()
	err := writeLogins(readLogins())
	loginsMu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	read()
	if parses != 1 {
		t.Fatalf("writing the accounts the file already holds made it parse again: %d parses", parses)
	}

	// a change of an account is still written
	ls := readLogins()
	if len(ls) == 0 {
		t.Fatal("no accounts to change")
	}
	ls[0].Plan = "changed"
	loginsMu.Lock()
	err = writeLogins(ls)
	loginsMu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(loginsPath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(got, []byte(`"plan": "changed"`)) {
		t.Fatalf("the change was not written: %s", got)
	}
}

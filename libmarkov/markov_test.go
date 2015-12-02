package markov

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func newm() (*Markov, error) {
	m, e := NewMarkov("testdata/testmarkov.db", "testmarkov")
	if e != nil {
		return nil, e
	}
	m.Open()
	return m, nil
}

func delm(m *Markov) error {
	var rete error
	e := m.Close()
	if e != nil {
		rete = fmt.Errorf("del: %s", e)
	}
	fs, e := filepath.Glob("testdata/testmarkov.db*")
	if e != nil {
		return fmt.Errorf("%s, %s", rete, e)
	}
	for i := range fs {
		e = os.Remove(fs[i])
		if e != nil {
			rete = fmt.Errorf("%s, %s", rete, e)
		}
	}
	return rete
}

func TestPopulateFromFile(t *testing.T) {
	m, e := newm()
	if e != nil {
		t.Fatal("newm:", e)
	}
	e = m.PopulateFromFile("testdata/lipsum.txt", true)
	if e != nil {
		t.Error(e)
	}
	e = delm(m)
	if e != nil {
		t.Error("delm", e)
	}
}

func TestChainmark(t *testing.T) {
	m, e := newm()
	if e != nil {
		t.Fatal("newm:", e)
	}
	e = m.PopulateFromFile("testdata/lipsum.txt", true)
	if e != nil {
		t.Fatal("PopulateFromFile", e)
	}
	s, e := m.Chainmark("lorem", 10, 5)
	if (e != nil) && (s == "") {
		t.Error("Chainmark:", e)
	} else if e != nil {
		t.Log("Chainmark:", e)
	}
	e = delm(m)
	if e != nil {
		t.Error("delm", e)
	}
}

func TestAddString(t *testing.T) {
	m, e := newm()
	if e != nil {
		t.Fatal("newm:", e)
	}
	f, e := os.Open("testdata/lipsum.txt")
	if e != nil {
		t.Fatal("open:", e)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for line, e := r.ReadString('\n'); e != io.EOF && e == nil; line, e = r.ReadString('\n') {
		e = m.AddString(line, true)
		if e != nil {
			t.Error("AddString:", e)
		}
	}
	s, e := m.Chainmark("lorem", 10, 5)
	if (e != nil) && (s == "") {
		t.Error("Chainmark:", e)
	} else if e != nil {
		t.Log("Chainmark:", e)
	}
	e = delm(m)
	if e != nil {
		t.Error("delm", e)
	}
}

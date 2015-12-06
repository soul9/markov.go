package markov

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

//because of sql i need to do a const dammit
const (
	Maxindex  = 10
	MaxWords  = 100
	commitlen = 5000
)

var (
	ErrNotEnoughWords = errors.New("Couldn't chain enough words")
	ErrNoWords        = errors.New("No words found")
	smartsep          = []rune{'.', '!', '?'}
)

type Markov struct {
	db        *sql.DB
	tablename string
	dbfile    string
}

func MarkSQLType() string {
	s := "(word TEXT,"
	for i := 1; i < Maxindex; i++ {
		s = fmt.Sprintf("%s idx%d,", s, i)
	}
	s = fmt.Sprintf("%s idx%d)", s, Maxindex)
	return s
}

type TableName string

func MarkSqlIndex() []func(TableName) string {
	r := make([]func(TableName) string, 0)
	for i := 1; i <= Maxindex; i++ {
		s := "("
		s1 := fmt.Sprintf("index%d", i)
		for j := i; j < Maxindex; j++ {
			s = fmt.Sprintf("%sidx%d, ", s, j)
		}
		s = fmt.Sprintf("%sidx%d)", s, Maxindex)
		f := func(t TableName) string {
			return fmt.Sprintf("CREATE INDEX IF NOT EXISTS '%s_%s' ON '%s' %s;", t, s1, t, s)
		}
		r = append(r, f)
	}
	return r
}

func NewMarkov(dbfile, tablename string) (*Markov, error) {
	m := &Markov{}
	m.tablename = tablename
	m.dbfile = dbfile
	e := m.Open()
	defer m.Close()
	if e == nil {
		_, e = m.db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS '%s' %s;", m.tablename, MarkSQLType()))
		if e != nil {
			return m, e
		}
		for _, f := range MarkSqlIndex() {
			_, e = m.db.Exec(f(TableName(tablename)))
			if e != nil {
				return m, e
			}
		}
	}
	return m, e
}

func (m *Markov) Open() error {
	var err error
	m.db, err = sql.Open("sqlite3", m.dbfile)
	return err
}

func (m *Markov) Close() error {
	e := m.db.Close()
	m.db = nil
	return e
}

func (m *Markov) Populate(toadd *bufio.Reader, smart bool) error {
	return Populate(m.db, m.tablename, toadd, smart)
}

func (m *Markov) AddString(toadd string, smart bool) error {
	return AddString(m.db, m.tablename, toadd, smart)
}

func (m *Markov) PopulateFromFile(fname string, smart bool) error {
	return PopulateFromFile(m.db, m.tablename, fname, smart)
}

func (m *Markov) Chainmark(s string, l int, idxno int) (string, error) {
	return Chainmark(m.db, m.tablename, s, l, idxno)
}

func prepareTx(db *sql.DB, qstr string) (*sql.Tx, *sql.Stmt, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, nil, err
	}
	st, err := tx.Prepare(qstr)
	if err != nil {
		st.Close()
		tx.Commit()
		return nil, nil, fmt.Errorf("Problem with sql statement: %s: %s", qstr, err)
	}
	return tx, st, nil
}

func trimsmart(r rune) bool {
	for _, s := range smartsep {
		if r == s {
			return true
		}
	}
	return false
}

func Populate(db *sql.DB, tablename string, toadd *bufio.Reader, smart bool) error {
	w := make([]interface{}, Maxindex+1)
	qstr := fmt.Sprintf("INSERT INTO '%s' (idx1", tablename)
	// idx2,idx3,idx4,idx5,idx6,idx7,idx8,idx9,idx10, word) values(?,?,?,?,?,?,?,?,?,?,?);"
	for i := 2; i <= Maxindex; i++ {
		qstr = fmt.Sprintf("%s, idx%d", qstr, i)
	}
	qstr = fmt.Sprintf("%s, word) values(?", qstr)
	for i := 1; i <= Maxindex; i++ {
		qstr = fmt.Sprintf("%s, ?", qstr)
	}
	qstr = fmt.Sprintf("%s);", qstr)
	for i := 0; i < len(w); i++ {
		w[i] = " "
	}
	commit := 0
	var tx *sql.Tx
	var st *sql.Stmt
	var err error
	for line, err := toadd.ReadString('\n'); err != io.EOF && err == nil; line, err = toadd.ReadString('\n') {
		if commit%commitlen == 0 {
			tx, st, err = prepareTx(db, qstr)
			if err != nil {
				return err
			}
		}
		commit++
		for _, ww := range strings.Split(line, " ") {
			if ww == "" {
				continue
			}
			w[len(w)-1] = strings.TrimFunc(strings.ToLower(strings.TrimSpace(ww)), trimsmart)
			_, err = st.Exec(w...)
			if err != nil {
				st.Close()
				e := tx.Commit()
				if e != nil {
					err = fmt.Errorf("%s, commit: %s", err, e)
				}
				return fmt.Errorf("Couldn't execute sql statement: %s: %s", qstr, err)
			}
			for i := 0; i < len(w)-1; i++ {
				w[i] = w[i+1]
			}
			if smart {
				sw := ww[len(ww)-1:]
				//Makes the algorithm a little bit "smarter". This makes it "see" phrases
				if sw == "." || sw == "!" || sw == "?" {
					for i := 0; i < Maxindex; i++ {
						w[i] = " "
					}
				}
			}
		}
		if commit%commitlen == 0 {
			st.Close()
			err = tx.Commit()
			if err != nil {
				return err
			}
		}
	}
	st.Close()
	e := tx.Commit()
	if e != nil {
		if err != nil {
			err = fmt.Errorf("%s, commit: %s", err, e)
		} else {
			err = e
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func AddString(db *sql.DB, tablename string, toadd string, smart bool) error {
	r := bufio.NewReader(strings.NewReader(toadd))
	err := Populate(db, tablename, r, smart)
	return err
}

func PopulateFromFile(db *sql.DB, tablename string, fname string, smart bool) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	err = Populate(db, tablename, r, smart)
	return err
}

func Chainmark(db *sql.DB, tablename string, s string, l int, idxno int) (string, error) {
	tidyret := func(s []string) string {
		return strings.TrimSpace(strings.Join(s, " "))
	}
	if idxno > Maxindex {
		return "", errors.New("Given index count is larger than the maximum allowable index")
	}
	if l > MaxWords {
		return "", errors.New("Too many words requested")
	}
	rand.Seed(time.Now().UnixNano())
	splitab := strings.Split(strings.TrimSpace(strings.ToLower(s)), " ")
	for i := range splitab {
		splitab[i] = strings.TrimFunc(splitab[i], trimsmart)
	}
	retab := make([]string, l+len(splitab))
	copy(retab, splitab)
	w := make([]string, idxno)
	for i := range w {
		w[i] = " "
	}
	if len(splitab) < idxno {
		for i, elm := range splitab {
			w[len(w)-i-1] = elm
			retab[i] = elm
		}
	} else {
		copy(w, splitab[len(splitab)-idxno:])
		copy(retab, splitab)
	}
	for i := len(splitab); i < l+len(splitab); i++ {
		qstr := fmt.Sprintf("from '%s' WHERE", tablename)
		empty := true
		tmpt := make(map[int]string)
		if w[0] != " " {
			qstr = fmt.Sprintf("%s idx%d=?", qstr, Maxindex-idxno+1)
			empty = false
			tmpt[len(tmpt)] = w[0]
		}
		for i := 1; i < idxno; i++ {
			if w[i] != " " {
				if !empty {
					qstr = fmt.Sprintf("%s AND", qstr)
				}
				qstr = fmt.Sprintf("%s idx%d=?", qstr, Maxindex-idxno+i+1)
				tmpt[len(tmpt)] = w[i]
				empty = false
			}
		}
		qstr = fmt.Sprintf("%s;", qstr)
		tmps := make([]interface{}, len(tmpt))
		for i := 0; i < len(tmpt); i++ {
			tmps[i] = tmpt[i]
		}
		st, err := db.Prepare(fmt.Sprintf("SELECT count(word) %s", qstr))
		if err != nil {
			return tidyret(retab), fmt.Errorf("Couldn't prepare statement: SELECT count(word) %s: %s", qstr, err)
		}
		res, err := st.Query(tmps...)
		if err != nil {
			st.Close()
			return tidyret(retab), fmt.Errorf("exec statement: SELECT count(word) %s: %s", qstr, err)
		}
		var cnt int
		if res.Next() {
			res.Scan(&cnt)
		}
		if cnt == 0 {
			res.Close()
			st.Close()
			switch tidyret(retab[len(splitab):]) {
			case "":
				return tidyret(retab), ErrNoWords
			default:
				return tidyret(retab), ErrNotEnoughWords
			}
		}
		res.Close()
		st.Close()
		st, err = db.Prepare(fmt.Sprintf("SELECT word %s", qstr))
		if err != nil {
			return tidyret(retab), fmt.Errorf("Couldn't prepare statement: SELECT word %s: %s", qstr, err)
		}
		res, err = st.Query(tmps...)
		if err != nil {
			res.Close()
			st.Close()
			return tidyret(retab), fmt.Errorf("exec statement: SELECT word %s: %s", qstr, err)
		}
		rnd := rand.Intn(cnt)
		var c string
		res.Next()
		for i := 0; i < rnd-1; i++ {
			if !res.Next() {
				res.Close()
				st.Close()
				return tidyret(retab), res.Err()
			}
		}
		for i := 0; i < idxno-1; i++ {
			w[i] = w[i+1]
		}
		res.Scan(&c)
		retab[i] = c
		w[len(w)-1] = retab[i]
		res.Close()
		st.Close()
	}
	return tidyret(retab), nil
}

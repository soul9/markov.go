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

const MarkSqlType = "(word TEXT, idx1 TEXT, idx2 TEXT, idx3 TEXT, idx4 TEXT, idx5 TEXT, idx6 TEXT, idx7 TEXT, idx8 TEXT, idx9 TEXT, idx10 TEXT)"

//because of sql i need to do a const dammit
const (
	Maxindex  = 10
	MaxWords  = 100
	commitlen = 5000
)

type Markov struct {
	db     *sql.DB
	dbname string
	dbfile string
}

func NewMarkov(dbfile, dbname string) (*Markov, error) {
	m := &Markov{}
	m.dbname = dbname
	m.dbfile = dbfile
	e := m.Open()
	defer m.Close()
	if e == nil {
		_, e = m.db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s %s;", m.dbname, MarkSqlType))
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
	m.Open()
	defer m.Close()
	e := Populate(m.db, m.dbname, toadd, smart)
	return e
}

func (m *Markov) AddString(toadd string, smart bool) error {
	m.Open()
	defer m.Close()
	e := AddString(m.db, m.dbname, toadd, smart)
	return e
}

func (m *Markov) PopulateFromFile(fname string, smart bool) error {
	m.Open()
	defer m.Close()
	e := PopulateFromFile(m.db, m.dbname, fname, smart)
	return e
}

func (m *Markov) Chainmark(s string, l int, idxno int) (error, string) {
	m.Open()
	defer m.Close()
	e, s := Chainmark(m.db, m.dbname, s, l, idxno)
	return e, s
}

func Populate(db *sql.DB, dbname string, toadd *bufio.Reader, smart bool) error {
	w := make([]interface{}, Maxindex+1)
	qstr := fmt.Sprintf("INSERT INTO %s (idx1", dbname)
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
	st, err := db.Prepare(qstr)
	if err != nil {
		return errors.New(fmt.Sprintf("Problem with sql statement: %s\n%s", qstr, err))
	}
	commit := 0
	_, err = db.Exec("BEGIN")
	if err != nil {
		return err
	}
	defer db.Exec("COMMIT")
	for line, err := toadd.ReadString('\n'); err != io.EOF && err == nil; line, err = toadd.ReadString('\n') {
		if commit%commitlen == 0 {
			_, err = db.Exec("BEGIN")
		}
		commit++
		for _, w[len(w)-1] = range strings.Split(line, " ") {
			if w[len(w)-1] == "" || w[len(w)-1] == " " {
				continue
			}
			w[len(w)-1] = strings.ToLower(strings.TrimSpace(w[len(w)-1].(string)))
			_, err = st.Exec(w...)
			if err != nil {
				return errors.New(fmt.Sprintf("Couldn't execute sql statement: %s\n(error was: %s", qstr, err))
			}
			for i := 0; i < len(w)-1; i++ {
				w[i] = w[i+1]
			}
			if smart {
				//Makes the algorithm a little bit "smarter". This makes it "see" phrases
				if len(w[len(w)-1].(string)) != 0 && (w[len(w)-1].(string)[:1] == "." || w[len(w)-1].(string)[:1] == "!" || w[len(w)-1].(string)[:1] == "?") {
					for i := 0; i < Maxindex; i++ {
						w[i] = " "
					}
				}
			}
		}
		if commit%commitlen == 0 {
			db.Exec("COMMIT")
		}
	}
	st.Close()
	return nil
}

func AddString(db *sql.DB, dbname string, toadd string, smart bool) error {
	r := bufio.NewReader(strings.NewReader(toadd))
	err := Populate(db, dbname, r, smart)
	return err
}

func PopulateFromFile(db *sql.DB, dbname string, fname string, smart bool) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	err = Populate(db, dbname, r, smart)
	return err
}

func Chainmark(db *sql.DB, dbname string, s string, l int, idxno int) (error, string) {
	if idxno > Maxindex {
		return errors.New("Given index count is larger than the maximum allowable index"), ""
	}
	if l > MaxWords {
		return errors.New("Too many words requested"), ""
	}
	rand.Seed(time.Now().UnixNano())
	splitab := strings.Split(strings.TrimSpace(strings.ToLower(s)), " ")
	retab := make([]string, l+len(splitab))
	copy(retab, splitab)
	w := make([]string, idxno)
	for i, _ := range w {
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
		qstr := fmt.Sprintf("from %s WHERE", dbname)
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
		for i := 0; i <= len(tmpt)-1; i++ {
			tmps[i] = tmpt[i]
		}
		st, err := db.Prepare(fmt.Sprintf("SELECT count(word) %s", qstr))
		if err != nil {
			return errors.New(fmt.Sprintf("Couldn't prepare statement: SELECT count(word) %s: %s", qstr, err)), strings.Join(retab, " ")
		}
		res, err := st.Query(tmps...)
		if err != nil {
			return errors.New(fmt.Sprintf("exec statement: SELECT count(word) %s\n%s", qstr, err)), strings.Join(retab, " ")
		}
		var cnt int
		if res.Next() {
			res.Scan(&cnt)
		}
		if cnt == 0 {
			return errors.New(fmt.Sprintf("Couldn't continue with this word combination: %v, %v, sql: select * %s", w, tmps, qstr)), strings.Join(retab, " ")
		}
		st.Close()
		st, err = db.Prepare(fmt.Sprintf("SELECT word %s", qstr))
		if err != nil {
			return errors.New(fmt.Sprintf("Couldn't prepare statement: SELECT word %s: %s", qstr, err)), strings.Join(retab, " ")
		}
		res, err = st.Query(tmps...)
		if err != nil {
			return errors.New(fmt.Sprintf("exec statement: SELECT word %s: %s", qstr, err)), strings.Join(retab, " ")
		}
		rnd := rand.Intn(cnt)
		var c string
		res.Next()
		for i := 0; i < rnd-1; i++ {
			if !res.Next() {
				return res.Err(), strings.Join(retab, " ")
			}
		}
		for i := 0; i < idxno-1; i++ {
			w[i] = w[i+1]
		}
		res.Scan(&c)
		retab[i] = c
		w[len(w)-1] = retab[i]
		st.Close()
	}
	return nil, strings.Join(retab, " ")
}

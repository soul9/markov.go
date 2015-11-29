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
	"strconv"
	"strings"
	"time"
)

const MarkSqlType = "(word TEXT, idx1 TEXT, idx2 TEXT, idx3 TEXT, idx4 TEXT, idx5 TEXT, idx6 TEXT, idx7 TEXT, idx8 TEXT, idx9 TEXT, idx10 TEXT)"

//because of sql i need to do a const dammit
const (
	maxindex  = 10
	commitlen = 5000
)

func Populate(db *sql.DB, dbname string, toadd *bufio.Reader, smart bool) error {
	w := make([]interface{}, maxindex+1)
	qstr := "INSERT INTO " + dbname + " (idx1"
	// idx2,idx3,idx4,idx5,idx6,idx7,idx8,idx9,idx10, word) values(?,?,?,?,?,?,?,?,?,?,?);"
	for i := 2; i <= maxindex; i++ {
		qstr = qstr + ", idx" + strconv.Itoa(i)
	}
	qstr = qstr + ", word) values(?"
	for i := 1; i <= maxindex; i++ {
		qstr = qstr + ", ?"
	}
	qstr = qstr + ");"
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
			if w[len(w)-1] == "" {
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
					for i := 0; i < maxindex; i++ {
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

func PopulateFromFile(db *sql.DB, dbname string, fname string, smart bool) error {
	if idxno > maxindex {
		return errors.New("Given index count is larger than the maximum allowable index")
	}
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	err = Populate(db, dbname, r, smart)
	if err != nil {
		return err
	}
	return nil
}

func Chainmark(db *sql.DB, dbname string, s string, l int, idxno int) (error, string) {
	if idxno > maxindex {
		return errors.New("Given index count is larger than the maximum allowable index"), ""
	}
	rand.Seed(time.Now().UnixNano())
	splitab := strings.Split(strings.ToLower(s), " ")
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
		qstr := "from " + dbname + " WHERE"
		empty := true
		tmpt := make(map[int]string)
		if w[0] != " " {
			qstr = qstr + " idx" + strconv.Itoa(maxindex-idxno+1) + "=?"
			empty = false
			tmpt[len(tmpt)] = w[0]
		}
		for i := 2; i <= idxno; i++ {
			if w[i-1] != " " {
				if !empty {
					qstr = qstr + " AND"
				}
				qstr = qstr + " idx" + strconv.Itoa(maxindex-idxno+i) + "=?"
				tmpt[len(tmpt)] = w[i-1]
				empty = false
			}
		}
		qstr = qstr + ";"
		tmps := make([]interface{}, len(tmpt))
		for i := 0; i <= len(tmpt)-1; i++ {
			tmps[i] = tmpt[i]
		}
		st, err := db.Prepare("SELECT count(word) " + qstr)
		if err != nil {
			return errors.New(fmt.Sprintf("Couldn't prepare statement: SELECT count(*) %s: %s", qstr, err)), strings.Join(retab, " ")
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
			return errors.New(fmt.Sprintf("Couldn't continue with this word combination:%v, %v, sql: select * %s", w, tmps, qstr)), strings.Join(retab, " ")
		}
		st.Close()
		st, err = db.Prepare("SELECT word " + qstr)
		if err != nil {
			return errors.New(fmt.Sprintf("Couldn't prepare statement: SELECT word %s", err)), strings.Join(retab, " ")
		}
		res, err = st.Query(tmps...)
		if err != nil {
			return errors.New("exec statement: SELECT count(*) " + qstr), strings.Join(retab, " ")
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

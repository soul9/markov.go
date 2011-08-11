package markov

import (
    "fmt"
    "os"
    "bufio"
    "strings"
    "rand"
    "time"
    sqlite "gosqlite.googlecode.com/hg/sqlite"
    "strconv"
)

const MarkSqlType = "(word TEXT, idx1 TEXT, idx2 TEXT, idx3 TEXT, idx4 TEXT, idx5 TEXT, idx6 TEXT, idx7 TEXT, idx8 TEXT, idx9 TEXT, idx10 TEXT)"
//because of sql i need to do a const dammit
const (
    maxindex = 10
    commitlen = 5000
)

func Populate(db *sqlite.Conn, dbname string, toadd *bufio.Reader, smart bool, idxno int) os.Error{
    w := make([]interface{}, maxindex+1)
    qstr := "INSERT INTO " + dbname + " (idx1"
    // idx2,idx3,idx4,idx5,idx6,idx7,idx8,idx9,idx10, word) values(?,?,?,?,?,?,?,?,?,?,?);"
    for i:=2; i<=maxindex; i++ {
        qstr = qstr + ", idx" + strconv.Itoa(i)
    }
    qstr = qstr + ", word) values(?"
    for i:=1; i<=maxindex; i++ {
        qstr = qstr + ", ?"
    }
    qstr = qstr + ");"
    for i:=0; i<len(w); i++ {
        w[i] = " "
    }
    st, err := db.Prepare(qstr)
    if err != nil {
        return os.NewError("Problem with sql statement: " + qstr + "\n" + err.String())
    }
    commit := 0
    err = db.Exec("BEGIN")
    defer db.Exec("COMMIT")
    for line, err := toadd.ReadString('\n'); err != os.EOF && err == nil; line, err = toadd.ReadString('\n') {
         if commit % commitlen == 0{
             err = db.Exec("BEGIN")
         }
         commit++
         for _, w[len(w)-1] = range strings.Split(line, " ") {
             if w[len(w)-1]  == "" {
                 continue
             }
             w[len(w)-1]  = strings.ToLower(strings.TrimSpace(w[len(w)-1].(string) ))
             err = st.Exec(w...)
             st.Next()
             if err != nil {
                 return os.NewError("Couldn't execute sql statement: " + qstr + "\n(error was: " + err.String())
             }
             for i := 0; i<len(w)-1; i++ {
                 w[i] = w[i+1]
             }
             if smart {
            //Makes the algorithm a little bit "smarter". This makes it "see" phrases
                 if len(w[len(w)-1].(string)) != 0 && ( w[len(w)-1].(string)[:1] == "." || w[len(w)-1].(string)[:1] == "!" || w[len(w)-1].(string)[:1] == "?" ) {
                     for i:=0; i<idxno; i++ {
                         w[i] = " "
                     }
                 }
             }
         }
         if commit % commitlen == 0 {
             db.Exec("COMMIT")
         }
    }
    st.Finalize()
    return nil
}

func PopulateFromFile(db *sqlite.Conn, dbname string, fname string, smart bool, idxno int) os.Error{
    if idxno > maxindex {
        return os.NewError("Given index count is larger than the maximum allowable index")
    }
    f, err := os.Open(fname)
    if err != nil {
        return err
    }
    defer f.Close()
    r := bufio.NewReader(f)
    err = Populate(db, dbname, r, smart, idxno)
    if err != nil {
        return err
    }
    return nil
}

func Chainmark(db *sqlite.Conn, dbname string, s string, l int, idxno int) (os.Error, string) {
    if idxno > maxindex {
        return os.NewError("Given index count is larger than the maximum allowable index"), ""
    }
    rand.Seed(time.Nanoseconds())
    splitab := strings.Split(strings.ToLower(s), " ")
    retab := make([]string, l+len(splitab))
    copy(retab, splitab)
    w:=make([]string, idxno)
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
    for i:=len(splitab); i < l + len(splitab); i++ {
        qstr := "from " + dbname + " WHERE"
        empty := true
        tmpt := make(map[int]string)
        if w[0] != " " {
            qstr = qstr + " idx" + strconv.Itoa(maxindex-idxno+1) + "=?"
            empty = false
            tmpt[len(tmpt)] = w[0]
        }
        for i:=2; i<=idxno; i++ {
            if w[i-1] != " " {
                if ! empty {
                    qstr = qstr + " AND"
                }
                qstr = qstr + " idx"+strconv.Itoa(maxindex-idxno+i)+"=?"
                tmpt[len(tmpt)] = w[i-1]
                empty = false
            }
        }
        qstr = qstr + ";"
        tmps := make([]interface{}, len(tmpt))
        for i:=0; i<=len(tmpt)-1; i++ {
            tmps[ i] = tmpt[i]
        }
        st, err := db.Prepare("SELECT count(word) "+qstr)
        if err != nil {
            return os.NewError(fmt.Sprintf("Couldn't prepare statement: SELECT count(*) %s: %s", qstr, err.String())), strings.Join(retab, " ")
        }
        err = st.Exec(tmps...)
        if err != nil {
            return os.NewError("exec statement: SELECT count(word) "+qstr+"\n" + err.String()), strings.Join(retab, " ")
        }
        var cnt int
        if st.Next() {
            st.Scan(&cnt)
        }
        if cnt == 0 {
            return os.NewError(fmt.Sprintf("Couldn't continue with this word combination:%v, %v, sql: select * %s", w, tmps, qstr)), strings.Join(retab, " ")
        }
        st.Finalize()
        st, err = db.Prepare("SELECT word "+qstr)
        if err != nil {
            return os.NewError("Couldn't prepare statement: SELECT word " + err.String()), strings.Join(retab, " ")
        }
        err = st.Exec(tmps...)
        if err != nil {
            return os.NewError("exec statement: SELECT count(*) "+qstr), strings.Join(retab, " ")
        }
        rnd := rand.Intn(cnt)
        var c string
        st.Next()
        for i:=0; i<rnd-1; i++ {
            if !st.Next() {
                return st.Error(), strings.Join(retab, " ")
            }
        }
        for i := 0; i<idxno-1; i++ {
             w[i] = w[i+1]
         }
        st.Scan(&c)
        retab[i] = c
        w[len(w)-1] = retab[i]
        st.Finalize()
    }
    return nil, strings.Join(retab, " ")
}
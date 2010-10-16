package main

import (
    "fmt"
    "os"
    "flag"
    "bufio"
    "strings"
    "rand"
    "time"
    sqlite "gosqlite.googlecode.com/hg/sqlite"
    "strconv"
)

const MarkSqlType = "(word TEXT, idx1 TEXT, idx2 TEXT, idx3 TEXT, idx4 TEXT, idx5 TEXT, idx6 TEXT, idx7 TEXT, idx8 TEXT, idx9 TEXT, idx10 TEXT)"
//because of sql i need to do a const dammit
const maxindex = 10;

type MarkStruct struct {
    idxs [10]string
    word string
}

type MarkTab map[string]map[string]map[int]string

func populate(db *sqlite.Conn, dbname string, fname string, smart bool, idxno int) os.Error{
    f, err := os.Open(fname, os.O_RDONLY, 0)
    if err != nil {
        return err
    }
    defer f.Close()
    r := bufio.NewReader(f)
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
    for line, err := r.ReadString('\n'); err != os.EOF && err == nil; line, err = r.ReadString('\n') {
         for _, w[len(w)-1] = range strings.Split(line, " ", -1) {
             if w[len(w)-1]  == "" {
                 continue
             }
             w[len(w)-1]  = strings.ToLower(strings.TrimSpace(w[len(w)-1].(string) ))
             st, err := db.Prepare(qstr)
             if err != nil {
                 return os.NewError("Problem with sql statement: " + qstr + "\n" + err.String())
             }
             err = st.Exec(w...)
             st.Next()
             if err != nil {
                 return os.NewError("Couldn't execute sql statement: " + qstr + "\n(error was: " + err.String())
             }
             st.Finalize()
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
    }
    return nil
}

func chainmark(db *sqlite.Conn, dbname string, s string, l int, idxno int) (os.Error, string) {
    rand.Seed(time.Nanoseconds())
    splitab := strings.Split(strings.ToLower(s), " ", -1)
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
            qstr = qstr + " idx" + strconv.Itoa(maxindex-idxno) + "=?"
            empty = false
            tmpt[len(tmpt)] = w[0]
        }
        for i:=1; i<=idxno; i++ {
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
        st, err := db.Prepare("SELECT count(word) "+qstr)
        tmps := make([]interface{}, len(tmpt))
        for i:=0; i<=len(tmpt)-1; i++ {
            tmps[i] = tmpt[i]
        }
        if err != nil {
            return os.NewError("Couldn't prepare statement: SELECT count(*) "+qstr), strings.Join(retab, " ")
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

func main () {
    fname := flag.String("c", "./eiro.log", "Corpus file path")
    startstring := flag.String("s", " ", "string to start with (defaults to space)")
    dbname := flag.String("n", "markov", "table name")
    dbfname := flag.String("d", "/tmp/testmarkov.sqlite3", "database file name")
    idxlen := flag.Int("i", 7, "number of indexes to use")
    smart := flag.Bool("m", false, "Smart mode: try and analyze test to detect sentences")
    retlen := flag.Int("l", 20, "How many words to chain")
    flag.Parse()
    fmt.Println(*fname, *startstring, *dbname, *dbfname, *idxlen, *smart, *retlen)
    if *idxlen > 10 {
        fmt.Println("Too many indexes, maximum is 10")
        os.Exit(1)
    }
    db, err := sqlite.Open(*dbfname)
    if err != nil {
        fmt.Println("Can't open database file.")
        os.Exit(1)
    }
    defer db.Close()
/*
    err = db.Exec("DROP TABLE IF EXISTS " + *dbname + ";")
    if err != nil {
        println(err);
    }

    err = db.Exec("CREATE TABLE " + *dbname + " "+ MarkSqlType + ";")
    if err != nil {
        fmt.Printf("Can't create table: %s\n%s", *dbname, err)
        os.Exit(1)
    }
    err = populate(db, *dbname, *fname, *smart, *idxlen)
    if err != nil {
        fmt.Printf("%s\n", err)
        os.Exit(1)
    }
*/
    err, str := chainmark(db, *dbname, *startstring, *retlen, *idxlen)
    if err != nil {
        fmt.Printf("Error in chainmark: %s\n", err)
    }
    fmt.Printf("%s\n", str)
    os.Exit(0)
}

package main

import (
    "fmt"
    "os"
    "flag"
    "bufio"
    "strings"
    "rand"
    "time"
    "sqlite3"
    "strconv"
)

const MarkSqlType = "(word TEXT, idx1 TEXT, idx2 TEXT, idx3 TEXT, idx4 TEXT, idx5 TEXT, idx6 TEXT, idx7 TEXT, idx8 TEXT, idx9 TEXT, idx10 TEXT)"

type MarkStruct struct {
    idxs [10]string
    word string
}

type MarkTab map[string]map[string]map[int]string

func populate(db *sqlite3.Handle, dbname string, fname string, smart bool, idxno int) os.Error{
    f, err := os.Open(fname, os.O_RDONLY, 0)
    if err != nil {
        return err
    }
    defer f.Close()
    r := bufio.NewReader(f)
    w := make([]string, idxno)
    qstr := "INSERT INTO " + dbname + " values (?,?,?,?,?,?,?,?,?,?,?);"
    for i:=0; i<idxno; i++ {
        w[i] = " "
    }
    for line, err := r.ReadString('\n'); err != os.EOF && err == nil; line, err = r.ReadString('\n') {
         for _, word := range strings.Split(line, " ", -1) {
             st, errstr := db.Prepare(qstr)
             if errstr != "" {
                 return os.NewError("Problem with sql statement: " + qstr + " " + errstr)
             }
             word = strings.ToLower(strings.TrimSpace(word))
             st.BindText(1, word)
             i := 0
             for ; i<idxno; i++ {
                 st.BindText(i+2, w[i])
             }
             for ; i<10; i++ {
                 st.BindText(i+2, " ")
             }
             err := st.Step()
             if err != 101 {
                 return os.NewError("Couldn't execute sql statement: " + db.ErrMsg() + " (error number: " + strconv.Itoa(err))
             }
             for i := 0; i<=idxno-2; i++ {
                 w[i] = w[i+1]
             }
             w[idxno-1] = word
              errint := st.Finalize()
              if errint != 0 {
                  return os.NewError("Couldn't close the database.\n" + db.ErrMsg())
              }
             if smart {
            //Makes the algorithm a little bit "smarter". This makes it "see" phrases
                 if len(word) != 0 && ( word[:1] == "." || word[:1] == "!" || word[:1] == "?" ) {
                     for i:=0; i<idxno; i++ {
                         w[i] = " "
                     }
                 }
             }
         }
    }
    return nil
}

func chainmark(db *sqlite3.Handle, dbname string, s string, l int, idxno int) (os.Error, string) {
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
            w[len(w)-1-i] = elm
            retab[i] = elm
        }
    } else {
        copy(w, splitab[len(splitab)-idxno:])
        copy(retab, splitab)
    }
    for i:=len(splitab); i < l + len(splitab); i++ {    
        qstr := "SELECT word from " + dbname + " WHERE"
        empty := true
        if w[0] != " " {
            qstr = qstr + " idx1=?"
            empty = false
        }
        for i:=2; i<=idxno; i++ {
            if w[i-1] != " " {
                if ! empty {
                    qstr = qstr + " AND"
                }
                qstr = qstr + " idx"+strconv.Itoa(i)+"=?"
                empty = false
            }
        }
        qstr = qstr + ";"
        st, err := db.Prepare(qstr)
        if err != "" {
            return os.NewError("Couldn't prepare statement: " + err), strings.Join(retab, " ")
        }
        sqlidx:=1
        for i:=0; i<idxno; i++ {
            if w[i] != " " {
                st.BindText(sqlidx, w[i])
                sqlidx++
            }
        }
        rnd := rand.Intn(st.ColumnCount())
        c := st.Step()
        for i:=0; i<rnd; i++ {
            c = st.Step()
        }
        switch c {
            case sqlite3.SQLITE_DONE: 
                //TODO: try and find an other random number that has
                //elements in the array
                return os.NewError("Couldn't continue with this word combination"), strings.Join(retab, " ")
            case sqlite3.SQLITE_ROW:
                break
            default:
                return os.NewError("Problem getting results: " +db.ErrMsg()), strings.Join(retab, " ")
        }
        retab[i] = st.ColumnText(0)
        for i := 0; i<idxno-1; i++ {
             w[i] = w[i+1]
         }
        w[idxno-1] = retab[i]
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
    smart := flag.Bool("m", true, "Smart mode: try and analyze test to detect sentences")
    retlen := flag.Int("l", 20, "How many words to chain")
    flag.Parse()
    fmt.Println(*fname, *startstring, *dbname, *dbfname, *idxlen, *smart, *retlen)
    if *idxlen > 10 {
        fmt.Println("Too many indexes, maximum is 10")
        os.Exit(1)
    }
    sqlite3.Initialize()
    defer sqlite3.Shutdown()
    db := new(sqlite3.Handle)
    defer db.Close()
    errstr := db.Open(*dbfname)
    if errstr != "" {
        fmt.Println("Can't open database file.")
        os.Exit(1)
    }
/*    st, errstr := db.Prepare("DROP TABLE IF EXISTS " + *dbname + ";")
    if errstr != "" {
        println(errstr);
    } else {
        st.Step()
        st.Finalize()
    }

    st, errstr = db.Prepare("CREATE TABLE " + *dbname + " "+ MarkSqlType + ";")
    if errstr != "" {
        fmt.Printf("Can't create table: %s\n%s", *dbname, errstr)
        os.Exit(1)
    } else {
        st.Step()
        st.Finalize()
    }
    err := populate(db, *dbname, *fname, *smart, *idxlen)
    if err != nil {
        fmt.Printf("%s\n", err)
        os.Exit(1)
    }
*/
    err, str := chainmark(db, *dbname, *startstring, *retlen, *idxlen)
    if err != nil {
        fmt.Printf("Error in chainmark: %s\n", err)
        os.Exit(1)
    }
    fmt.Printf("%s\n", str)
    os.Exit(0)
}

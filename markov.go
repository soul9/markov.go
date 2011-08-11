package main

import (
    "fmt"
    "os"
    "flag"
    sqlite "gosqlite.googlecode.com/hg/sqlite"
    markov "github.com/soul9/markov.go/libmarkov"
)

func main () {
    fname := flag.String("c", "./eiro.log", "Corpus file path")
    startstring := flag.String("s", " ", "string to start with (defaults to space)")
    dbname := flag.String("n", "markov", "table name")
    dbfname := flag.String("d", "/tmp/testmarkov.sqlite3", "database file name")
    idxlen := flag.Int("i", 7, "number of indexes to use")
    smart := flag.Bool("m", false, "Smart mode: try and analyze test to detect sentences")
    retlen := flag.Int("l", 20, "How many words to chain")
    pop := flag.Bool("p", false, "Whether to populate the database or not")
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
    if *pop {
        err = db.Exec("DROP TABLE IF EXISTS " + *dbname + ";")
        if err != nil {
            println(err);
        }
    
        err = db.Exec("CREATE TABLE " + *dbname + " "+ markov.MarkSqlType + ";")
        if err != nil {
            fmt.Printf("Can't create table: %s\n%s", *dbname, err)
            os.Exit(1)
        }
        err = markov.PopulateFromFile(db, *dbname, *fname, *smart, *idxlen)
        if err != nil {
            fmt.Printf("%s\n", err)
            os.Exit(1)
        }
    }
    err, str := markov.Chainmark(db, *dbname, *startstring, *retlen, *idxlen)
    if err != nil {
        fmt.Printf("Error in chainmark: %s\n", err)
    }
    fmt.Printf("%s\n", str)
    os.Exit(0)
}

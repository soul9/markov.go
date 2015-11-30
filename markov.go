package main

import (
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	markov "github.com/soul9/markov.go/libmarkov"
	"os"
)

func main() {
	fname := flag.String("c", "none", "Corpus file path")
	startstring := flag.String("s", " ", "string to start with (defaults to space)")
	dbname := flag.String("n", "markov", "table name")
	dbfname := flag.String("d", "/tmp/testmarkov.sqlite3", "database file name")
	idxlen := flag.Int("i", 7, "number of indexes to use")
	smart := flag.Bool("m", false, "Smart mode: try and analyze test to detect sentences")
	retlen := flag.Int("l", 20, "How many words to chain")
	pop := flag.Bool("p", false, "Whether to populate the database or not")
	flag.Parse()
	fmt.Println(*fname, *startstring, *dbname, *dbfname, *idxlen, *smart, *retlen)
	if *idxlen > markov.Maxindex {
		fmt.Printf("Too many indexes, maximum is %d\n", markov.Maxindex)
		os.Exit(1)
	}
	m, err := markov.NewMarkov(*dbfname, *dbname)
	if err != nil {
		fmt.Println("Can't open database file:", err)
		os.Exit(1)
	}
	if *pop {
		err = m.PopulateFromFile(*fname, *smart)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
	}
	err, str := m.Chainmark(*startstring, *retlen, *idxlen)
	if err != nil {
		fmt.Printf("Error in chainmark: %s\n", err)
	}
	fmt.Printf("%s\n", str)
	os.Exit(0)
}

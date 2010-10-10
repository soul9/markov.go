package main

import (
    "fmt"
    "os"
    "flag"
    "bufio"
    "strings"
    "rand"
    "time"
)

type MarkTab map[string]map[string]map[int]string

func populate(t MarkTab, fname string) os.Error{
    f, err := os.Open(fname, os.O_RDONLY, 0)
    if err != nil {
        return err
    }
    defer f.Close()
    r := bufio.NewReader(f)
    w1 := " "
    w2 := " "
    for line, err := r.ReadString('\n'); err != os.EOF && err == nil; line, err = r.ReadString('\n') {
         for _, word := range strings.Split(line, " ", -1) {
             word = strings.ToLower(strings.TrimSpace(word) )
             if w1 != " " {
                 w1 = strings.ToLower(strings.TrimSpace(w1))
             }
             if w2 != " " {
                 w2 = strings.ToLower(strings.TrimSpace(w2))
             }
             if _, ok := t[w1]; !ok {
                 t[w1] = make(map[string]map[int]string)
             }
             if _, ok := t[w1][w2]; !ok {
                 t[w1][w2] = make(map[int]string)
             }
             t[w1][w2][len(t[w1][w2])] = word
             w1, w2 = w2, word
//Makes the algorithm a little bit "smarter". This makes it "see" phrases
             if len(word) != 0 && ( word[:1] == "." || word[:1] == "!" || word[:1] == "?" ) {
                 w1, w2 = " ", " "
             }
         }
    }
    if _, ok := t[w1]; !ok {
        t[w1] = make(map[string]map[int]string)
        t[w1][w2] = make(map[int]string)
    }
    if _, ok := t[w1][w2]; !ok {
        t[w1][w2] = make(map[int]string)
    }
    t[w1][w2][len(t[w1][w2])] = "\n"
    if err != os.EOF {
        return err
    }
    return nil
}

func chainmark(t MarkTab, s string, l int) (os.Error, string) {
    rand.Seed(time.Nanoseconds())
    w1, w2, w := " ", " ", " "
    strtab := make(map[int]string)
    for i, w := range strings.Split(s, " ", -1) {
        strtab[i] = w
    }
    if len(strtab) >=1 {
        w2 = strings.ToLower(strtab[len(strtab)-1])
        if len(strtab) >=2 {
            w1 = strings.ToLower(strtab[len(strtab)-2])
        }
    }
    for i:=0; i < l; i++ {
        if _, ok := t[w1]; !ok {
            //try and find an other random number that has
            //elements in the array
            i := 0
            max := rand.Intn(len(t[w1]))
            for w1, _ = range t {
                if i == max && len(t[w1]) != 0 {
                    break
                }
                i++
            }
        }
        if _, ok := t[w1][w2]; !ok {
            //try and find an other random number that has
            //elements in the array
            i := 0
            max := rand.Intn(len(t[w1]))
            for w2, _ = range t[w1] {
                if i == max && len(t[w1][w2]) != 0 {
                    break
                }
                i++
            }
        }
        rnd := rand.Intn(len(t[w1][w2]) )
        w = t[w1][w2][rnd]
        strtab[len(strtab)] = w
        w1, w2 = w2, w
    }
    ret := strtab[0]
    for i := 1; i<len(strtab); i++ {
        ret = ret + " " + strtab[i]
    }
    return nil, ret
}

func main () {
    fname := flag.String("c", "./eiro.log", "Corpus file path")
    startstring := flag.String("s", "", "string to start with (defaults to space)")
    flag.Parse()
    mtab := make(MarkTab)
    err := populate(mtab, *fname)
    if err != nil {
        fmt.Printf("%s\n", err)
        os.Exit(1)
    }
    err, str := chainmark(mtab, *startstring, 100)
    fmt.Printf("%s\n", str)
    os.Exit(0)
}

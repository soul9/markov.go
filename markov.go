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

func populate(t *MarkTab, fname string) (os.Error, MarkTab) {
    f, err := os.Open(fname, os.O_RDONLY, 0)
    ret := *t
    if err != nil {
        return err, ret
    }
    defer f.Close()
    r := bufio.NewReader(f)
    w1 := " "
    w2 := " "
    for line, err := r.ReadString('\n'); err != os.EOF && err == nil; line, err = r.ReadString('\n') {
         for _, word := range strings.Split(line, " ", -1) {
             word = strings.ToLower(strings.TrimSpace(word))
             if w1 != " " {
                 w1 = strings.ToLower(strings.TrimSpace(w1))
             }
             if w2 != " " {
                 w2 = strings.ToLower(strings.TrimSpace(w2))
             }
             var tmpt map[string]map[int]string
             var tmpt2 map[int]string
             tmpt, ok := ret[w1]
             if !ok {
                 tmpt = make(map[string]map[int]string)
                 ret[w1] = tmpt
             }
             tmpt2, ok = tmpt[w2]
             if !ok {
                 tmpt2 = make(map[int]string)
             }
             tmpt2[len(tmpt2)] = word
             ret[w1][w2] = tmpt2
             w1, w2 = w2, word
         }
    }
    if _, ok := ret[w1]; !ok {
        ret[w1] = make(map[string]map[int]string)
        ret[w1][w2] = make(map[int]string)
    }
    if _, ok := ret[w1][w2]; !ok {
        ret[w1][w2] = make(map[int]string)
    }
    ret[w1][w2][len(ret[w1][w2])] = "\n"
    if err != os.EOF {
        return err, ret
    }
    return nil, ret
}

func chainmark(t *MarkTab, s string, l int) (os.Error, string) {
    rand.Seed(time.Nanoseconds() % 1e9)
    w1, w2, w := " ", " ", " "
    strtab := strings.Split(s, " ", -1)
    if len(strtab) >=1 {
        w2 = strings.ToLower(strtab[len(strtab)-1])
        if len(strtab) >=2 {
            w1 = strings.ToLower(strtab[len(strtab)-2])
        }
    }
    for i:=0; i < l; i++ {
        if len((*t)[w1]) == 0 {
            continue    //first iteration, w2 needs to get filled
        }
        if len((*t)[w1][w2])  == 0 {
            //try and find an other random number that has
            //elements in the array
            i := 0
            max := rand.Intn(len((*t)[w1]))
            for w2, _ = range (*t)[w1] {
                w2 = strings.ToLower(w2)
                if i == max && len((*t)[w1][w2]) != 0 {
                    break
                }
                i++
            }
        }
        newtab := make([]string, len(strtab)+1)
        copy(newtab, strtab)
        strtab = newtab
        rnd := rand.Intn(len((*t)[w1][w2]) )
        w = strings.ToLower((*t)[w1][w2][rnd])
        strtab[len(strtab)-1] = w
        w1, w2= w2, w
    }
    return nil, strings.Join(strtab, " ")
}

func main () {
    fname := flag.String("c", "-c /path/to/corpus.txt", "Corpus file path")
    startstring := flag.String("s", "-s 'start string'", "string to start with (defaults to space)")
    flag.Parse()
    mtab := make(MarkTab)
    err, tab := populate(&mtab, *fname)
    if err != nil {
        fmt.Printf("%s\n", err)
    }
    err, str := chainmark(&tab, *startstring, 50)
    fmt.Printf("%s\n", str)
}

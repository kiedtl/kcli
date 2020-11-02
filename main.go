// TODO: hungarian (apps) notation

package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	//"os"
	"regexp"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// TODO: arg parsing!
var (
	enmul  int
	dbpath string
)

func main() {
	enmul = 25
	dbpath = "./kcli.db"

	// open database
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	var input string
	fmt.Scan(&input)
	words := regexp.MustCompile("[\\n\\ ~!@#$%^&*,.?]+").Split(input, -1)

	context := ""

	learn(words, db)
	talk(words, context, db)
}

func talk(stuff []string, context string, db *sql.DB) {
	noun := get_noun(stuff, context, db)
	nonsense := generate(noun, db)

	fmt.Println(strings.Join(nonsense, " "))
}

func get_noun(stuff []string, context string, db *sql.DB) string {
	nouns := column_from_table(db, "word", "noun")

	out := count_dups(nouns)
	noun := least_duplicated(out)

	ins_conv, err := db.Prepare("INSERT OR IGNORE INTO conver (pre, pro) VALUES (?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	if context != "" {
		ins_conv.Exec(context, noun)
	}

	conv_q, err := db.Query("SELECT pro FROM conver WHERE pre=?", noun)
	if err != nil {
		log.Fatal(err)
	}

	defer conv_q.Close()

	var nextnouns []string
	var item string
	for conv_q.Next() {
		err := conv_q.Scan(&item)
		if err != nil {
			log.Fatal(err)
		}

		nextnouns = append(nextnouns, item)
	}

	if len(nextnouns) > 0 {
		noun = nextnouns[rand.Intn(len(nextnouns))]
	}

	return noun
}

func generate(noun string, db *sql.DB) []string {
	beg := column_from_table(db, "word", "beg")
	end := column_from_table(db, "word", "end")
	nouns := column_from_table(db, "word", "noun")

	iter := 0
	out := []string{noun}

	prew_q_beg, err := db.Query("SELECT pre FROM prew WHERE pro=?", out[0])
	if err != nil {
		log.Fatal(err)
	}

	defer prew_q_beg.Close()

	var beg_words []string
	var beg_word string
	for prew_q_beg.Next() {
		err := prew_q_beg.Scan(&beg_word)
		if err != nil {
			log.Fatal(err)
		}

		beg_words = append(beg_words, beg_word)
	}

	for (!contains(out[0], beg) || (count_occurs(out[0], nouns)-1) > iter*enmul) && iter < 7 {
		if len(beg_words) > 0 {
			newword := []string{beg_words[rand.Intn(len(beg_words))]}
			out = append(newword, out...)
		} else {
			break
		}

		iter += 1
	}

	prew_q_end, err := db.Query("SELECT pro FROM prew WHERE pre=?", out[len(out)-1])
	if err != nil {
		log.Fatal(err)
	}

	defer prew_q_end.Close()

	var end_words []string
	var end_word string
	for prew_q_end.Next() {
		err := prew_q_end.Scan(&end_word)
		if err != nil {
			log.Fatal(err)
		}

		end_words = append(end_words, end_word)
	}

	for (!contains(out[len(out)-1], end) || (count_occurs(out[len(out)-1], nouns)-1) > iter*enmul) && iter < 7 {
		if len(end_words) > 0 {
			newword := []string{end_words[rand.Intn(len(end_words))]}
			out = append(newword, out...)
		} else {
			break
		}

		iter += 1
	}

	return out
}

func learn(stuff []string, db *sql.DB) {
	var pre string

	// TODO: create tables if they do not exist

	ins_beg, err := db.Prepare("INSERT INTO beg (word) VALUES (?)")
	if err != nil {
		log.Fatal(err)
	}

	ins_prew, err := db.Prepare("INSERT OR IGNORE INTO prew (pre, pro) VALUES (?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	ins_noch, err := db.Prepare("INSERT INTO noun (word) VALUES (?)")
	if err != nil {
		log.Fatal(err)
	}

	ins_end, err := db.Prepare("INSERT INTO end (word) VALUES (?)")
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range stuff {
		if pre == "" {
			// insert word into beg
			ins_beg.Exec(item)
		} else {
			// insert pre, word into prew
			ins_prew.Exec(pre, item)
		}

		pre = item
		// insert word into noch
		ins_noch.Exec(item)
	}

	// insert pre into end
	ins_end.Exec(pre)
}

func count_occurs(value string, data []string) int {
	cnt := 0
	for _, item := range data {
		if item == value {
			cnt += 1
		}
	}

	return cnt
}

func count_dups(data []string) map[string]int {
	dup_freq := make(map[string]int)
	for _, item := range data {
		_, in_map := dup_freq[item]
		if in_map {
			dup_freq[item] += 1
		} else {
			dup_freq[item] = 1
		}
	}

	return dup_freq
}

func least_duplicated(data map[string]int) string {
	type kv struct {
		Key   string
		Value int
	}

	var ss []kv
	for k, v := range data {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	return ss[0].Key
}

// return <column> of all rows from table
// TODO: just return error instead of printing it
func column_from_table(db *sql.DB, column string, table string) []string {
	// it's OK to be using Sprintf here because
	// column and table are not user inputs
	stmt := fmt.Sprintf("SELECT %s FROM %s", column, table)
	qq, err := db.Query(stmt)
	if err != nil {
		log.Fatal(err)
	}

	defer qq.Close()

	var data []string
	var item string
	for qq.Next() {
		err := qq.Scan(&item)
		if err != nil {
			log.Fatal(err)
		}

		data = append(data, item)
	}

	return data
}

func contains(value string, array []string) bool {
	for _, v := range array {
		if value == v {
			return true
		}
	}

	return false
}

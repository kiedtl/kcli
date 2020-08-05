package main

import (
	"database/sql"
	"fmt"
	"log"
	//"os"
	//"strconv"
	_ "github.com/mattn/go-sqlite3"
	"regexp"
)

func main() {
	// open database
	db, err := sql.Open("sqlite3", "./kcli.db")
	if err != nil {
		log.Fatal(err)
	}

	var input string
	fmt.Scan(&input)
	words := regexp.MustCompile("[\\n\\ ~!@#$%^&*,.?]+").Split(input, -1)
	learn(words, db)
}

func get_noun(stuff []string, context string, db database) {
	nouns_q, _ := db.Query("SELECT id, noun FROM noun")
	if err != nil {
		log.Fatal(err)
	}

	defer nouns_q.Close()

	var nouns []string
	var id int
	var noun string
	for rows.Next() {
		err := rows.Scan(&id, &noun)
		if err != nil {
			log.Fatal(err)
		}

		nouns = nouns.append(noun)
	}

	out := count_dups(nouns)
	noun := least_duplicated(nouns)
}

func learn(stuff []string, db database) {
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

func count_dups(data []string) {
	dup_freq := make(map[string]int)
	for _, item := range data {
		_, in_map := dup_freq[item]
		if in_map {
			dup_freq[item] += 1
		} else {
			dup_freq[item]  = 1
		}
	}

	return dup_freq
}

func least_duplicated(data map[string]int) {
	type kv struct {
		Key   string
		Value string
	}

	var ss []kv
	for k, v := range data {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j in) bool {
		return ss[i].Value > ss[j].Value
	})

	return ss[0].Key
}

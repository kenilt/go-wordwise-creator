package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
)

const (
	WordwiseDictionaryPath = "resources/wordwise-dict-optimized.csv"
	LemmaDictionaryPath    = "resources/lemmatization-en.csv"
)

type DictRow struct {
	Word    string
	Phoneme string
	En      string
	Vi      string
	HintLvl int
}

// Load Dict from CSV
func loadWordwiseDict() *map[string]DictRow {

	file, err := os.Open(WordwiseDictionaryPath)
	if err != nil {
		log.Fatalln("Error when open ", WordwiseDictionaryPath, "->", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	reader := csv.NewReader(file)

	dict := make(map[string]DictRow)
	row := DictRow{}

	// Read each record from csv
	// skip header
	record, err := reader.Read()
	if err == io.EOF {
		log.Fatalln("Empty csv file")
	}

	count := 0

	for {
		record, err = reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("Error when scan word ", count, "->", err)
		}

		if len(record) < 5 {
			log.Println("Invalid word: ", record)
			continue
		}

		hintLv, err := strconv.Atoi(record[4])
		if err != nil {
			log.Println("Can't get hint_level: ", record, "->", err)
			continue
		}

		row = DictRow{
			Word:    record[0],
			Phoneme: record[1],
			En:      record[2],
			Vi:      record[3],
			HintLvl: hintLv,
		}

		dict[row.Word] = row
		count++
	}

	log.Println("--> Csv words:", count)
	return &dict
}

// Load Dict from CSV
func loadLemmatizerDict() *map[string]string {

	file, err := os.Open(LemmaDictionaryPath)
	if err != nil {
		log.Fatalln("Error when open ", LemmaDictionaryPath, "->", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	reader := csv.NewReader(file)

	dict := make(map[string]string)

	var record []string
	count := 0
	for {
		record, err = reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("Error when scan word ", count, "->", err)
		}

		if len(record) < 2 {
			log.Println("Invalid word: ", record)
			continue
		}

		dict[record[1]] = record[0]
		count++
	}

	log.Println("--> Lemma pairs:", count)
	return &dict
}

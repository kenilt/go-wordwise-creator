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

var wordwiseDict *map[string]DictRow
var lemmaDict *map[string]string

type DictRow struct {
	Word    string
	Phoneme string
	En      string
	Vi      string
	HintLvl int
}

func (ws *DictRow) meaning(isVietnamese bool) string {
	if isVietnamese {
		if len(ws.Phoneme) > 0 {
			return ws.Phoneme + " " + ws.Vi
		} else {
			return ws.Vi
		}
	} else {
		return ws.En
	}
}

// Load Dict from CSV
func loadWordwiseDict() {

	file, err := os.Open(WordwiseDictionaryPath)
	if err != nil {
		logFatalln("Error when open ", WordwiseDictionaryPath, "->", err)
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
		logFatalln("Empty csv file")
	}

	count := 0

	for {
		record, err = reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logFatalln("Error when scan word ", count, "->", err)
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
	wordwiseDict = &dict
}

// Load Dict from CSV
func loadLemmatizerDict() {

	file, err := os.Open(LemmaDictionaryPath)
	if err != nil {
		logFatalln("Error when open ", LemmaDictionaryPath, "->", err)
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
			logFatalln("Error when scan word ", count, "->", err)
		}

		if len(record) < 2 {
			log.Println("Invalid word: ", record)
			continue
		}

		dict[record[1]] = record[0]
		count++
	}

	log.Println("--> Lemma pairs:", count)
	lemmaDict = &dict
}

package main

import (
	"encoding/csv"
	"log"
	"os"
	"fmt"
	"strings"

	"bufio"
	"text/tabwriter"
	"flag"
)

var delimiter = ';'
var quote = "\""

func main() {

	parseManually := flag.Bool("parse-manually", false, "parse the first csv manually and forgiving")
	skipMissing := flag.Bool("skip-missing", false, "skip counting missing")
	skipDifferences := flag.Bool("skip-differences", false, "skip recording differences")
	onlyDifferenceColumn := flag.Int("diff-column", 0, "only record differences in column (0=all)")
	del := flag.Int("delimiter", 0, "the delimiter used (0=semilcolon, 1=tabulator, 2=comma)")
	flag.Parse()

	switch *del {
	case 0:
			delimiter = ';'
		case 1:
			delimiter = '\t'
		case 2:
			delimiter = ','
	}

	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.AlignRight|tabwriter.Debug)

	filename1 := flag.Arg(0)
	filename2 := flag.Arg(1)

	var recordsFile1 [][]string;
	if *parseManually {
		recordsFile1 = readCSVManually(filename1)
	}	else {
		recordsFile1 = readCSV(filename1)
	}
	recordsFile2 := readCSV(filename2)

	if !*skipMissing {
		fmt.Fprintln(w, "gesamt Datei 1\t", len(recordsFile1))
		fmt.Fprintln(w, "gesamt Datei 2\t", len(recordsFile2))

		missingfile1 := findMissing(recordsFile1, recordsFile2)
		fmt.Fprintln(w, "fehlend in Datei 1 aus Datei 2\t", len(missingfile1) - 1)
		writeCSV("missing_in_" + filename1, missingfile1)

		missingfile2 := findMissing(recordsFile2, recordsFile1)
		fmt.Fprintln(w, "fehlend in Datei 2 aus Datei 1\t", len(missingfile2) - 1)
		writeCSV("missing_in_" + filename2, missingfile2)
	}

	if !*skipDifferences {
		differences := findDifferences(recordsFile1, recordsFile2, *onlyDifferenceColumn)
		fmt.Fprintln(w, "Anzahl Datensatzabweichungen\t", (len(differences) - 1) / 2)
		writeCSV("differences.csv", differences)
	}

	w.Flush()
}

func readCSV(name string) [][]string {
	file, err := os.Open(name)
	if err != nil {
		panic(err)
	}

	fr := bufio.NewReader(file)

	r := csv.NewReader(fr)
	r.Comma = ';'
	r.Comment = '#'

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	return records
}

func readCSVManually(name string) [][]string {
	file, err := os.Open(name)
	if err != nil {
		panic(err)
	}

	records := make([][]string, 0)

	s := bufio.NewScanner(file)
	for s.Scan() {
		parts := strings.Split(s.Text(), string(delimiter))

		for i := 0; i < len(parts); i++ {
			if strings.Contains(parts[i], quote) {
				p := strings.Trim(parts[i], quote)
				parts[i] = strings.Trim(p, " ")
			}

		}

		records = append(records, parts)
	}
	err = s.Err()
	if err != nil {
		log.Fatal(err)
	}

	return records
}

func writeCSV(name string, records [][]string) {
	file, err := os.Create(name)
	if err != nil {
		panic(err)
	}

	fw := bufio.NewWriter(file)

	w := csv.NewWriter(fw)
	w.Comma = ';'
	w.WriteAll(records)

	if err := w.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
}

func findMissing(records [][]string, reference [][]string) [][]string {

	missingRecords := make([][]string, 0)
	missingRecords = append(missingRecords, reference[0])

	for i := 0; i < len(reference); i++ {
		m := reference[i]
		for j := 0; j < len(records); j++ {
			if records[j][0] == reference[i][0] {
				m = nil
				continue
			}
		}

		if m != nil {
			missingRecords = append(missingRecords, m)
		}
	}

	return missingRecords
}

func findDifferences(recordsA [][]string, recordsB [][]string, onlyDifferenceColumn int) [][]string {

	if onlyDifferenceColumn > 0 {
		differingRecords := make([][]string, 0)

		h := make([]string, 0)
		h = append(h, recordsA[0][0])
		h = append(h, recordsA[0][onlyDifferenceColumn])
		differingRecords = append(differingRecords, h)

		for i := 0; i < len(recordsA); i++ {
			for j := 0; j < len(recordsB); j++ {
				if recordsA[i][0] == recordsB[j][0] {
					if recordsA[i][onlyDifferenceColumn] != recordsB[j][onlyDifferenceColumn] {

						a := make([]string, 0)
						a = append(a, recordsA[i][0])
						a = append(a, recordsA[i][onlyDifferenceColumn])
						differingRecords = append(differingRecords, a)

						b := make([]string, 0)
						b = append(b, recordsB[j][0])
						b = append(b, recordsB[j][onlyDifferenceColumn])
						differingRecords = append(differingRecords, b)
					}
					continue
				}
			}
		}

		return differingRecords
	}


	differingRecords := make([][]string, 0)
	differingRecords = append(differingRecords, recordsA[0])

	for i := 0; i < len(recordsA); i++ {
		for j := 0; j < len(recordsB); j++ {
			if recordsA[i][0] == recordsB[j][0] {
				if differing(recordsA[i], recordsB[j]) {
					differingRecords = append(differingRecords, recordsA[i])
					differingRecords = append(differingRecords, recordsB[j])
				}
				continue
			}
		}
	}

	return differingRecords
}

func differing(recA []string, recB []string) bool {
	for i := 0; i < len(recA); i++ {
		if recA[i] != recB[i] {
		 return true
		}
	}
	return false
}

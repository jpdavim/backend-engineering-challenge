package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func Test_main_TemplateFile(t *testing.T) {

	data := getContentFromConsole()

	var expectedAverageDurationFirstEntry = 0.0
	var expectedAverageDurationLastEntry = 100.0

	if data[0].Average_delivery_time != expectedAverageDurationFirstEntry {
		t.Errorf("Expected duration for first minute of sample got %f, expected %f", data[0].Average_delivery_time, expectedAverageDurationFirstEntry)
	}

	if data[len(data)-1].Average_delivery_time != expectedAverageDurationLastEntry {
		t.Errorf("Expected duration for last minute of sample got %f, expected %f", data[len(data)-1].Average_delivery_time, expectedAverageDurationLastEntry)
	}

	var dateFirstEntry, _ = time.Parse("2006-01-02 15:04:05", data[0].Date)
	var dateLastEntry, _ = time.Parse("2006-01-02 15:04:05", data[len(data)-1].Date)

	// adding 1 to the number of entries account for the
	// entry for the last minute que got for rounding
	var numberOfEntries = int(dateLastEntry.Sub(dateFirstEntry).Minutes()) + 1

	if len(data) != numberOfEntries {
		t.Errorf("Expected number of minutes calculated got %d, expected %d", len(data), numberOfEntries)
	}

	for i := 23; i < 29; i++ {
		if data[i].Average_delivery_time != 0.0 {
			t.Errorf("Expected duration for %d minute of sample got %f, expected 0", i+1, data[i].Average_delivery_time)
		}
	}
}

func getContentFromConsole() []PrintableValues {

	os.Args = append(os.Args, "--input_file=./events-template.json")
	os.Args = append(os.Args, "--window_size=10")

	getStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	consoleContentRaw, _ := io.ReadAll(r)
	os.Stdout = getStdout

	var deliveredTranslation []PrintableValues
	// the content we get from the console is in byte and is a series of json objects, not an array of json objects
	// 1 - convert it to a string
	// 2 - replace the line breaks with commas, thus separating the objects
	// 3 - trim the last comma since it makes the array not valid
	// 4 - encapsulate the entire string in brackets making it a valid array of PrintableValues objects
	consoleContent := strings.ReplaceAll(string(consoleContentRaw), "\n", ",")
	consoleContent = "[" + strings.TrimSuffix(consoleContent, ",") + "]"

	// with the above treatment, the
	err := json.Unmarshal([]byte(consoleContent), &deliveredTranslation)

	if err != nil {
		fmt.Println(err)
	}

	return deliveredTranslation
}

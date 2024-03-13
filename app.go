/*
	go-challenge solves a technical challenge for Unbabel

	Calculates the moving average for the time it took to deliver translations to clients.
	Receives two optional flags, if the flags are not present, it will use the default values.
	After performing the calculations, the program will output to the console.

	Usage:

	go-challenge [flags]

	The flags are

	--input-file
	Path to the file with the translations delivery's data.
	If the path is not valid, or it is unable to open the file the program will exit with an error.
	The default value is "./events.json".

	--window_size
	Positive integer with the width of the time window (in minutes) used to calculate the moving average.
	If the value is not a integer greater or equal to 0 the program will exit with an error.
	The default value is 10.
*/

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
)

// struct with the information read from file
// the file has more information, but since it is not needed it won't be loaded into memory
// Timestamp: minute the translations were delivered
// Duration: duration of the delivery
type DeliveredTranslation struct {
	Timestamp string `json:"timestamp"`
	Duration  int    `json:"duration"`
}

// struct with the calculated values to print
// CurrentMinute: minute in time to which we are making the calculations
// AverageDuration: average time it took to deliver translations in this minute
type PrintableValues struct {
	Date                  string  `json:"date"`
	Average_delivery_time float64 `json:"average_delivery_time"`
}

func main() {
	// define the flags and the default values
	filePath := flag.String("input_file", "./events.json", "path to the input file")
	windowSize := flag.Uint("window_size", 10, "window size used to calculate the moving average")
	flag.Parse()

	// call the function that will read the file and return the data from the file ready to perform the calculations
	translationsDeliveriesData, firstMinute, lastMinute := readTranslationsFileAndProcessData(*filePath)

	// this array will work as a FIFO/Queue to store the values of the moving window
	var movingAverageQueue []int

	// iterating from the first minute a delivery occurred to the last minute a delivery ocurred
	// using time.Time to progress in time
	for currentMinute := firstMinute; !currentMinute.After(lastMinute); currentMinute = currentMinute.Add(time.Minute) {
		var currentAverage float64

		// getting the duration of the deliveries for this minute in time
		// need to convert to string to use as a key in the map
		var currentMinuteData = translationsDeliveriesData[currentMinute.Format("2006-01-02 15:04:05")]

		// update the elements in the queue
		// if we don't have data for the current minute in the map, it defaults to 0
		movingAverageQueue = updateMovingWindowQueue(movingAverageQueue, *windowSize, currentMinuteData)

		// calculating the moving average
		currentAverage = calculateMovingAverage(movingAverageQueue)

		// create the object with the data to print
		printableValues, _ := json.Marshal(PrintableValues{
			Date:                  currentMinute.Format("2006-01-02 15:04:05"),
			Average_delivery_time: currentAverage,
		})

		// print the values to the console
		// the challenge mentions an output file, but not a name for the file
		// I'm also assuming some automated tests will be ran and the output will be read from the console
		fmt.Println(string(printableValues))
	}
}

// function to update the moving average queue
// encapsulates the logic to add and remove elements to/from the queue
func updateMovingWindowQueue(movingAverageQueue []int, windowSize uint, currentMinuteData int) []int {
	// add the current minute data to the FIFO
	movingAverageQueue = append(movingAverageQueue, currentMinuteData)

	// if the FIFO has more elements than the "windowSize" we remove the first element
	if int64(len(movingAverageQueue)) > int64(windowSize) {
		movingAverageQueue = movingAverageQueue[1:]
	}

	return movingAverageQueue
}

// function to calculate the moving average for the current window
func calculateMovingAverage(movingAverageQueue []int) float64 {
	var sum int
	var numberMinutesWithDeliveries = 0

	// cycle through the queue that holds the values for the current and past minutes within the window size interval
	for i := 0; i < len(movingAverageQueue); i++ {
		// this condition is necessary to be compliant with the example given that excludes minutes with no deliveries from the calculations
		if movingAverageQueue[i] > 0 {
			// calculate the sum for all values bigger than 0
			// calculate how many values bigger than 0 there are in que queue
			sum += movingAverageQueue[i]
			numberMinutesWithDeliveries++
		}

	}

	// guarding against the case that the file has in interval larger than the window size
	// in that case the default value is 0
	// else we divide the sum per the number of minutes with deliveries
	if numberMinutesWithDeliveries == 0 {
		return 0
	} else {
		return float64(sum) / float64(numberMinutesWithDeliveries)
	}
}

// function
// a map that for which minute in which translations were delivered has the sum of the duration of the deliveries
// the first minute a translation delivery occurred
// the last minute a translation delivery occurred
func readTranslationsFileAndProcessData(filePath string) (map[string]int, time.Time, time.Time) {

	// open the file using the path received in the command line flag
	file, error := os.Open(filePath)

	// exit with error if unable to open the file
	if error != nil {
		panic(error)
	}

	// defer the close of the file at the return of this function
	defer file.Close()

	var scanner = bufio.NewScanner(file)
	var firstMinute time.Time
	var deliveredTranslation DeliveredTranslation
	var numberTranslationsPerMinuteUTC = make(map[string]int)

	// read the file line by line
	for scanner.Scan() {

		// read the file and map the content to a DeliveredTranslation struct
		json.Unmarshal([]byte(scanner.Text()), &deliveredTranslation)

		// parsing the string timestamp to a time.Time object
		// truncating it to the minute - to have simpler keys in the map
		// adding one minute to the event - to make it coherent with the example
		// converting it back to a string
		currentMinute, _ := time.Parse("2006-01-02 15:04:05", deliveredTranslation.Timestamp)
		currentMinute = currentMinute.Truncate(time.Minute).Add(time.Minute)
		deliveredTranslation.Timestamp = currentMinute.Format("2006-01-02 15:04:05")

		// for each minute we had a delivery we calculate how long the deliveries for that minute took
		// and store them in a map whose key is the truncated timestamp - just the minute
		numberTranslationsPerMinuteUTC[deliveredTranslation.Timestamp] = numberTranslationsPerMinuteUTC[deliveredTranslation.Timestamp] + deliveredTranslation.Duration

		// since the information is stored in a map and not ordered
		// as the file is read the minute of the first event is stored
		if firstMinute.IsZero() {
			firstMinute, _ = time.Parse("2006-01-02 15:04:05", deliveredTranslation.Timestamp)
			firstMinute = firstMinute.Add(-time.Minute)
		}
	}

	// the last minute when a delivery ocurred is also stored
	lastMinute, _ := time.Parse("2006-01-02 15:04:05", deliveredTranslation.Timestamp)

	// return the values
	return numberTranslationsPerMinuteUTC, firstMinute, lastMinute
}

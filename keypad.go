package main

import (
	"fmt"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
	"os"
	"syscall"
	"time"
)

var chip *gpiod.Chip

var rowOffsets = []int{rpi.J8p26, rpi.J8p24, rpi.J8p23, rpi.J8p22}
var colOffsets = []int{rpi.J8p21, rpi.J8p19, rpi.J8p10}

var rowPins = [4]*gpiod.Line{}
var colPins = [3]*gpiod.Line{}

var keypad = [][]string{
	{"1", "2", "3"},
	{"4", "5", "6"},
	{"7", "8", "9"},
	{"*", "0", "#"},
}

func main() {
	var err error

	chip, err = gpiod.NewChip("gpiochip0")
	if err != nil {
		panic(err)
	}

	// Assign all row pins
	for i := 0; i <= len(rowOffsets) - 1; i++ {
		assignRow(rowOffsets[i])
	}


	// Assign all col pins
	for i := 0; i <= len(colOffsets) - 1; i++ {
		assignCol(colOffsets[i])
	}

	// Clean up
	defer closeLines()
	defer chip.Close()

	// Infinite loop waiting for inputs
	inputLoop()
}

func assignRow(offset int) {
	fmt.Println("[LOG] Assigning row pin", offset)

	l, err := chip.RequestLine(offset,
		gpiod.WithPullDown,
		gpiod.WithBothEdges)
	if err != nil {
		fmt.Printf("RequestLine returned error: %s\n", err)
		if err == syscall.Errno(22) {
			fmt.Println("Note that the WithPullUp option requires kernel V5.5 or later - check your kernel version.")
		}
		os.Exit(1)
	}

	// Add line to row pins array
	rowPins[indexOf(offset, rowOffsets)] = l
}

func assignCol(offset int) {
	fmt.Println("[LOG] Assigning col pin", offset)

	l, err := chip.RequestLine(offset,
		gpiod.AsOutput(1))

	if err != nil {
		fmt.Printf("RequestLine returned error: %s\n", err)
		os.Exit(1)
	}

	// Add line to col pins array
	colPins[indexOf(offset, colOffsets)] = l
}

func closeLines() {
	for i := 0; i <= len(rowPins) - 1; i++ {
		rowPins[i].Close()
	}

	for i := 0; i >= len(colPins) - 1; i++ {
		rowPins[i].Close()
	}
}

func indexOf(value int, array []int) int {
	for i := 0; i <= len(array) - 1; i++ {
		if array[i] == value {
			return i
		}
	}

	return -1
}

func inputLoop() {

	for true {
		time.Sleep(time.Millisecond * 10)

		var rowIndex int
		var rowLine *gpiod.Line

		// Attempt to find a pressed row
		for i := 0; i <= len(rowPins) - 1; i++ {
			// Get row value
			value, _ := rowPins[i].Value()

			// Check if row is pressed
			if value == 1 {
				rowIndex = i
				rowLine = rowPins[i]

				break
			}
		}

		// If row line not found
		if rowLine == nil {
			continue
		}

		// Set default selectedColValue
		selectedColIndex := -1

		// Check for pressed column
		for i := 0; i <= len(colPins) - 1; i++ {
			// Temporarily set col output to 0 to test if that's the one
			colPins[i].SetValue(0)

			// Read row rowValue
			rowValue, _ := rowLine.Value()

			// Set output back to 1
			colPins[i].SetValue(1)

			// If rowValue is now 0 then we found our row
			if rowValue == 0 {
				selectedColIndex = i

				break
			}
		}

		if selectedColIndex == -1 {
			fmt.Println("Something went wrong and no selected col was found.")
		}

		fmt.Println(keypad[rowIndex][selectedColIndex])

		// Wait for the button release before checking new input
		for true {
			state, _ := rowPins[rowIndex].Value()

			if state == 0 {
				break
			}

			time.Sleep(time.Millisecond * 10)
		}
	}

}
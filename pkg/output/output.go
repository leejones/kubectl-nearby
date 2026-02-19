package output

import (
	"fmt"
	"strings"
	"time"
)

func Age(duration time.Duration) string {
	if duration.Seconds() < 120 {
		return fmt.Sprintf("%vs", duration.Seconds())
	} else if duration.Seconds() < 600 {
		return duration.String()
	} else if duration.Minutes() < 120 {
		return fmt.Sprintf("%.0fm", duration.Truncate(time.Minute).Minutes())
	} else if duration.Hours() < 24 {
		return fmt.Sprintf("%.0fh", duration.Truncate(time.Hour).Hours())
	} else {
		return fmt.Sprintf("%.0fd", duration.Truncate(time.Hour).Hours()/24)
	}
}

func Columns(input [][]string) (string, error) {
	columnLengths := []int{}
	columnCount := len(input[0])
	for i := 0; i < columnCount; i++ {
		columnLengths = append(columnLengths, 0)
	}
	output := []string{}
	for _, row := range input {
		for index, item := range row {
			currentColumnLength := columnLengths[index]
			if currentColumnLength < len(item) {
				columnLengths[index] = len(item)
			}
		}
	}
	for _, row := range input {
		outputRow := []string{}
		for index, item := range row {
			columnLength := columnLengths[index]
			outputItem := item
			// Right pad item output with spaces unless it's the last column.
			if index != len(row)-1 {
				for len(outputItem) < columnLength {
					outputItem += " "
				}
			}
			outputRow = append(outputRow, outputItem)
		}
		output = append(output, strings.Join(outputRow, "  "))
	}
	return strings.Join(output, "\n"), nil
}

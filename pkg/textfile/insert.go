package textfile

import (
	"fmt"
	"strings"
)

// package for text operations

func InsertLine(code string, lineNumber int, newText string) (string, error) {
	if lineNumber < 1 {
		return "", fmt.Errorf("line number must be greater than or equal to 1")
	}

	lines := strings.Split(code, "\n")
	if lineNumber > len(lines)+1 {
		return "", fmt.Errorf("line number is greater than the number of lines in the code")
	}

	lines = append(lines[:lineNumber-1], append([]string{newText}, lines[lineNumber-1:]...)...)
	return strings.Join(lines, "\n"), nil
}

func InsertLinesAtIndices(file string, indices []int, linesToInsert []string) (string, error) {
	if len(indices) != len(linesToInsert) {
		return "", fmt.Errorf("the length of indices and linesToInsert must be the same")
	}

	lines := strings.Split(file, "\n")

	for i, index := range indices {
		if index < 1 || index > len(lines)+1 {
			return "", fmt.Errorf("index %d is out of bounds", index)
		}

		lines = append(lines[:index-1+i], append([]string{linesToInsert[i]}, lines[index-1+i:]...)...)
	}

	return strings.Join(lines, "\n"), nil
}

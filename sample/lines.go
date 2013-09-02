package main

import (
	"bufio"
	"fmt"
	"os"
)

func saveLines(lines []string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := bufio.NewWriter(file)
	defer buf.Flush()

	for _, line := range lines {
		if _, err := fmt.Fprintln(buf, line); err != nil {
			return err
		}
	}
	return nil
}

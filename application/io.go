package main

import (
	"bufio"
	"os"
	"strings"
)

func readLine() (string, ierrori) {

	var (
		line string
		e    error
	)

	line, e = bufio.NewReader(os.Stdin).ReadString('\n')
	if e != nil {
		return line, ierror{m: "Could not read line", e: e}
	}

	line = strings.Trim(line, " \n\r\t")

	return line, nil
}

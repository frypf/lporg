package utils

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
)

var (
	normalPadding    = cli.Default.Padding
	doublePadding    = normalPadding * 2
	triplePadding    = normalPadding * 3
	quadruplePadding = normalPadding * 4
)

// Indent indents apex log line
func Indent(f func(s string)) func(string) {
	return func(s string) {
		cli.Default.Padding = doublePadding
		f(s)
		cli.Default.Padding = normalPadding
	}
}

// DoubleIndent double indents apex log line
func DoubleIndent(f func(s string)) func(string) {
	return func(s string) {
		cli.Default.Padding = triplePadding
		f(s)
		cli.Default.Padding = normalPadding
	}
}

// TripleIndent triple indents apex log line
func TripleIndent(f func(s string)) func(string) {
	return func(s string) {
		cli.Default.Padding = quadruplePadding
		f(s)
		cli.Default.Padding = normalPadding
	}
}

// StringInSlice finds string in array
func StringInSlice(str string, list []string) bool {
	for _, entry := range list {
		if entry == str {
			return true
		}
	}
	return false
}

// AppendIfMissing appends a string to an array if it doesn't already contain that string
func AppendIfMissing(list []string, str string) []string {
	for _, entry := range list {
		if entry == str {
			return list
		}
	}
	return append(list, str)
}

// Unique removes duplicate numbers from an int array
func Unique(list []int) []int {
	keys := make(map[int]bool)
	newList := []int{}
	for _, entry := range list {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			newList = append(newList, entry)
		}
	}
	return newList
}

func checkError(err error) {
	if err != nil {
		log.WithError(err).Fatal("failed")
	}
}

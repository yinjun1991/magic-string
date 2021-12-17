package magicstring

import (
	"path/filepath"
	"regexp"
	"strings"
)

func getLocator(source string) func(index int) (line, column int) {
	originalLines := strings.Split(source, "\n")
	lineOffsets := make([]int, len(originalLines))

	for i, pos := 0, 0; i < len(originalLines); i++ {
		lineOffsets[i] = pos
		pos += len(originalLines[i]) + 1
	}

	return func(index int) (line, column int) {
		start := 0
		end := len(originalLines)
		for start < end {
			mid := (start + end) >> 1
			if index < lineOffsets[mid] {
				end = mid
			} else {
				start = mid + 1
			}
		}

		line = start - 1
		column = index - lineOffsets[line]

		return line, column
	}
}

var regPathSep = regexp.MustCompile("[/\\]")

func getRelativePath(from, to string) string {
	from = filepath.Dir(from)
	froms := regPathSep.Split(from, -1)
	tos := regPathSep.Split(to, -1)

	for froms[0] == tos[0] {
		froms = froms[1:]
		tos = tos[1:]
	}

	if len(froms) > 0 {
		for i := range froms {
			froms[i] = ".."
		}
	}

	froms = append(froms, tos...)
	return strings.Join(froms, "/")
}

var regTabIndent = regexp.MustCompile("^\t+")
var regSpaceIndent = regexp.MustCompile("^ {2,}")
var regSpaceStart = regexp.MustCompile("^ +")

func guessIndent(content string) string {
	lines := strings.Split(content, "\n")

	tabbed := 0
	spaced := make([][]byte, 0, len(lines))
	for _, line := range lines {
		lineBytes := []byte(line)
		if regTabIndent.Match(lineBytes) {
			tabbed++
		} else if regSpaceIndent.Match(lineBytes) {
			spaced = append(spaced, lineBytes)
		}
	}

	if tabbed > len(spaced) {
		return "\t"
	}

	if len(spaced) == 0 {
		return ""
	}

	minSpaces := []byte{}
	for _, lineBytes := range spaced {
		spaces := regSpaceStart.Find(lineBytes)
		if len(spaces) < len(minSpaces) {
			minSpaces = spaces
		}
	}

	return string(minSpaces)
}

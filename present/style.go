// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"html"
	"html/template"
	"strings"
	"unicode"
	"unicode/utf8"
)

/*
	Fonts are demarcated by an initial and final char bracketing a
	space-delimited word, plus possibly some terminal punctuation.
	The chars are
		_ for italic
		* for bold
		` (back quote) for fixed width.
	Inner appearances of the char become spaces. For instance,
		_this_is_italic_!
	becomes
		<i>this is italic</i>!
*/

func style(s string) template.HTML {
	return template.HTML(font(html.EscapeString(s)))
}

// font returns s with font indicators turned into HTML font tags.
func font(s string) string {
	if strings.IndexAny(s, "`_*") == -1 {
		return s
	}
	words := split(s)
	var b bytes.Buffer
Word:
	for w, word := range words {
		if len(word) < 2 {
			continue Word
		}
		const punctuation = `.,;:()!?—–'"`
		const marker = "_*`"
		// Initial punctuation is OK but must be peeled off.
		first := strings.IndexAny(word, marker)
		if first == -1 {
			continue Word
		}
		// Is the marker prefixed only by punctuation?
		for _, r := range word[:first] {
			if !strings.ContainsRune(punctuation, r) {
				continue Word
			}
		}
		open, word := word[:first], word[first:]
		char := word[0] // ASCII is OK.
		close := ""
		switch char {
		default:
			continue Word
		case '_':
			open += "<i>"
			close = "</i>"
		case '*':
			open += "<b>"
			close = "</b>"
		case '`':
			open += "<code>"
			close = "</code>"
		}
		// Terminal punctuation is OK but must be peeled off.
		last := strings.LastIndex(word, word[:1])
		if last == 0 {
			continue Word
		}
		head, tail := word[:last+1], word[last+1:]
		for _, r := range tail {
			if !strings.ContainsRune(punctuation, r) {
				continue Word
			}
		}
		b.Reset()
		b.WriteString(open)
		var wid int
		for i := 1; i < len(head)-1; i += wid {
			var r rune
			r, wid = utf8.DecodeRuneInString(head[i:])
			if r != rune(char) {
				// Ordinary character.
				b.WriteRune(r)
				continue
			}
			if head[i+1] != char {
				// Inner char becomes space.
				b.WriteRune(' ')
				continue
			}
			// Doubled char becomes real char.
			// Not worth worrying about "_x__".
			b.WriteByte(char)
			wid++ // Consumed two chars, both ASCII.
		}
		b.WriteString(close) // Write closing tag.
		b.WriteString(tail)  // Restore trailing punctuation.
		words[w] = b.String()
	}
	return strings.Join(words, "")
}

// split is like strings.Fields but also returns the runs of spaces.
func split(s string) []string {
	words := make([]string, 0, 10)
	prevWasSpace := false
	mark := 0
	for i, r := range s {
		isSpace := unicode.IsSpace(r)
		if isSpace != prevWasSpace && i > mark {
			words = append(words, s[mark:i])
			mark = i
		}
		prevWasSpace = isSpace
	}
	if mark < len(s) {
		words = append(words, s[mark:])
	}
	return words
}

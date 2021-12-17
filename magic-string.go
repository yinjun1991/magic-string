package magicstring

import (
	"errors"
	"fmt"
	"strings"
)

type ExclusionRange = [2]int

type MagicString struct {
	original              string
	intro, outro          string
	firstChunk, lastChunk *chunk
	lastSearchedChunk     *chunk
	byStart, byEnd        map[int]*chunk
	locator               func(index int) (line, column int)
}

// Overwrite the characters from start to end with content.
// The same restrictions as s.remove() apply.
// Return the current magic string.
func (ms *MagicString) Overwrite(start, end int, content string) (*MagicString, error) {
	for start < 0 {
		start += len(ms.original)
	}

	for end < 0 {
		end += len(ms.original)
	}

	if end > len(ms.original) {
		return nil, errors.New("end is out of bounds")
	}

	if start == end {
		return nil, errors.New("cannot overwrite a zero-length range, use AppendLeft or PrependRight instead")
	}

	ms.split(start)
	ms.split(end)

	firstC := ms.byStart[start]
	lastC := ms.byEnd[end]

	if firstC != nil {
		if end > firstC.end && firstC.next != ms.byStart[firstC.end] {
			return nil, errors.New("cannot overwrite across a split chunk")
		}

		firstC.edit(content)
		if firstC != lastC {
			for c := firstC.next; c != lastC; c = c.next {
				c.edit("")
			}
			lastC.edit("")
		}
	} else {
		c := newChunk(start, end, "")
		c.edit("")
		lastC.next = c
		c.prev = lastC
	}

	return ms, nil
}

func (ms *MagicString) String() string {
	strs := make([]string, 0, len(ms.byStart)+2)
	strs = append(strs, ms.intro)

	for c := ms.firstChunk; c != nil; c = c.next {
		strs = append(strs, c.String())
	}

	strs = append(strs, ms.outro)

	return strings.Join(strs, "")
}

// Prepend the string with the specified content.
// Return the current magic string.
func (ms *MagicString) Prepend(prefix string) *MagicString {
	return ms
}

// PrependLeft is the same as s.AppendLeft(...), except that the inserted content will go before any previous appends or prepends at index
// Return the current magic string.
func (ms *MagicString) PrependLeft(index int, content string) *MagicString {
	return ms
}

// PrependRight is the same as s.AppendRight(...), except that the inserted content will go before any previous appends or prepends at index
// Return the current magic string.
func (ms *MagicString) PrependRight(index int, content string) *MagicString {
	return ms
}

// Append the specified content to the end of the string.
// Return the current magic string.
func (ms *MagicString) Append(content string) *MagicString {
	ms.outro += content
	return ms
}

func (ms *MagicString) split(index int) {
	c := ms.byStart[index]
	if c != nil {
		return
	}
	c = ms.byEnd[index]
	if c != nil {
		return
	}

	c = ms.lastSearchedChunk
	search_forward := index > c.end

	for c != nil {
		if c.contains(index) {
			ms.splitChunk(c, index)
			return
		}
		if search_forward {
			c = ms.byStart[c.end]
		} else {
			c = ms.byEnd[c.start]
		}
	}
}

func (ms *MagicString) splitChunk(c *chunk, index int) error {
	if c.edited && len(c.content) > 0 {
		line, column := ms.locator(index)
		return fmt.Errorf("cannot split a chunk that has already been edited %d:%d", line, column)
	}
	c2 := c.split(index)
	ms.byEnd[index] = c
	ms.byStart[index] = c2
	ms.byEnd[c2.end] = c2
	if c == ms.lastChunk {
		ms.lastChunk = c2
	}
	ms.lastSearchedChunk = c
	return nil
}

// AppendLeft appends the specified content at the index in the original string.
// If a range ending with index is subsequently moved, the insert will be moved with it.
// Return the current magic string. See also s.PrependLeft(...).
func (ms *MagicString) AppendLeft(index int, content string) *MagicString {
	ms.split(index)
	c := ms.byEnd[index]

	if c != nil {
		c.appendLeft(content)
	} else {
		ms.intro += content
	}

	return ms
}

// AppendRight appends the specified content at the index in the original string.
// If a range starting with index is subsequently moved, the insert will be moved with it.
// Return the current magic string. See also s.PrependRight(...).
func (ms *MagicString) AppendRight(index int, content string) *MagicString {
	ms.split(index)
	c := ms.byStart[index]

	if c != nil {
		c.appendRight(content)
	} else {
		ms.outro += content
	}

	return ms
}

// Move the characters from start and end to index.
// Return the current magic string.
func (ms *MagicString) Move(start, end int, index int) error {
	if index >= start && index <= end {
		return errors.New("cannot move a selection inside itself")
	}

	ms.split(start)
	ms.split(end)
	ms.split(index)

	firstC := ms.byStart[start]
	lastC := ms.byEnd[end]

	oldLeft := firstC.prev
	oldRight := lastC.next

	newRight := ms.byStart[index]

	if newRight == nil && lastC == ms.lastChunk {
		return nil
	}

	var newLeft *chunk
	if newRight != nil {
		newLeft = newRight.prev
	} else {
		newLeft = ms.lastChunk
	}

	// connect the broken linked list
	if oldLeft != nil {
		oldLeft.next = oldRight
	}
	if oldRight != nil {
		oldRight.prev = oldLeft
	}

	firstC.prev = newLeft
	if newLeft != nil {
		newLeft.next = firstC
	} else {
		ms.firstChunk = firstC
	}

	lastC.next = newRight
	if newRight != nil {
		newRight.prev = lastC
	} else {
		ms.lastChunk = lastC
	}

	return nil
}

// Remove the characters from start to end (of the original string, not the generated string).
// Removing the same content twice, or making removals that partially overlap, will cause an error.
// Return the current magic string.
func (ms *MagicString) Remove(start, end int) error {
	for start < 0 {
		start += len(ms.original)
	}
	for end < 0 {
		end += len(ms.original)
	}

	if start == end {
		return nil
	}

	if end > len(ms.original) {
		return errors.New("out of bounds")
	}

	if start > end {
		return errors.New("end must be greater than start")
	}

	ms.split(start)
	ms.split(end)

	for c := ms.byStart[start]; c != nil; {
		c.intro = ""
		c.outro = ""
		c.edit("")

		if end > c.end {
			c = c.next
		} else {
			c = nil
		}
	}

	return nil
}

func New(input string) *MagicString {
	c := newChunk(0, len(input), input)

	ms := &MagicString{
		original:          input,
		firstChunk:        c,
		lastChunk:         c,
		lastSearchedChunk: c,
		byStart: map[int]*chunk{
			0: c,
		},
		byEnd: map[int]*chunk{
			len(input): c,
		},
		locator: getLocator(input),
	}

	return ms
}

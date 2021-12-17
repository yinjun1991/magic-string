package magicstring

type chunk struct {
	start, end   int
	original     string
	intro, outro string
	content      string
	edited       bool
	prev, next   *chunk
}

func (c *chunk) appendLeft(content string) {
	c.outro += content
}

func (c *chunk) appendRight(content string) {
	c.intro += content
}

func (c *chunk) prependLeft(content string) {
	c.outro = content + c.outro
}

func (c *chunk) prependRight(content string) {
	c.intro = content + c.intro
}

func (c *chunk) contains(index int) bool {
	return c.start < index && index < c.end
}

func (c *chunk) eachNext(fn func(*chunk)) {
	for c2 := c; c2 != nil; c2 = c.next {
		fn(c2)
	}
}

func (c *chunk) eachPrev(fn func(*chunk)) {
	for c2 := c; c2 != nil; c2 = c.prev {
		fn(c2)
	}
}

func (c *chunk) edit(content string) {
	c.content = content
	c.intro = ""
	c.outro = ""
	c.edited = true
}

// split one chunk into two chunks, return the back one.
// index MUST be in [c.start, c.end)
func (c *chunk) split(index int) *chunk {
	sliceIndex := index - c.start
	originBefore := c.original[0:sliceIndex]
	originAfter := c.original[sliceIndex:]

	c.original = originBefore

	c2 := newChunk(index, c.end, originAfter)
	c2.outro = c.outro
	c.outro = ""

	c.end = index

	if c.edited {
		c2.edit("")
		c.content = ""
	} else {
		c.content = originBefore
	}

	c2.next = c.next
	if c2.next != nil {
		c2.next.prev = c2
	}
	c2.prev = c
	c.next = c2
	return c2
}

func (c *chunk) String() string {
	return c.intro + c.content + c.outro
}

func newChunk(start, end int, content string) *chunk {
	return &chunk{
		start:    start,
		end:      end,
		original: content,
		content:  content,
	}
}

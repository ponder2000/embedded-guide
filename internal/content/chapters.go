// Package content holds the chapter registry for the visual guide.
//
// Each chapter is an HTML partial under web/templates/chapters/. To add a
// chapter you do two things: drop a new partial file, and add one entry to the
// Chapters slice below. Order in the slice defines reading order; prev/next
// links and the table of contents are derived from it automatically.
package content

// Chapter describes a single chapter of the guide.
type Chapter struct {
	// Num is the 1-based global chapter number, assigned from slice order.
	Num int
	// Slug is the URL segment: /chapter/<slug>/. Must be unique.
	Slug string
	// Title is the human-readable chapter title (may contain inline <em>).
	Title string
	// Summary is a one-line description shown in the table of contents.
	Summary string
	// Part groups chapters into sections on the home page.
	Part string
	// Template is the partial template name (defined via {{ define }}).
	Template string
	// Interactive flags chapters that contain a hands-on component, surfaced
	// with a small badge in the table of contents.
	Interactive bool
}

// Part groups chapters for display on the home page.
type Part struct {
	Name     string
	Chapters []Chapter
}

// Chapters is the ordered list of every chapter in the guide.
var Chapters = []Chapter{
	{Slug: "what-is-embedded", Title: "What <em>is</em> an embedded system?", Summary: "The anatomy of a microcontroller, and how it differs from the computer on your desk.", Part: "Foundations", Template: "chapter-what-is-embedded"},
	{Slug: "bits-and-bytes", Title: "Bits, bytes, hex &amp; <em>two's complement</em>", Summary: "How numbers really live in memory — and why negative numbers wrap around.", Part: "Foundations", Template: "chapter-bits-and-bytes"},
	{Slug: "bitwise-and-registers", Title: "Bitwise ops &amp; <em>registers</em>", Summary: "Set, clear and toggle individual bits on a live 32-bit hardware register.", Part: "Foundations", Template: "chapter-bitwise-and-registers", Interactive: true},

	{Slug: "memory-map", Title: "The memory map &amp; <em>memory-mapped I/O</em>", Summary: "Flash vs RAM, and why writing to an address can blink an LED.", Part: "The Machine", Template: "chapter-memory-map"},
	{Slug: "boot-sequence", Title: "Boot &amp; <em>startup</em>", Summary: "Step through the reset vector and the .data / .bss initialisation dance.", Part: "The Machine", Template: "chapter-boot-sequence", Interactive: true},
	{Slug: "clocks", Title: "Clocks, oscillators &amp; <em>prescalers</em>", Summary: "Drag a prescaler and watch the derived peripheral frequency change.", Part: "The Machine", Template: "chapter-clocks", Interactive: true},

	{Slug: "gpio", Title: "GPIO — <em>talking to a pin</em>", Summary: "Push-pull, open-drain and pull-ups, with a pin you can toggle yourself.", Part: "Talking to the World", Template: "chapter-gpio", Interactive: true},
	{Slug: "interrupts", Title: "Interrupts &amp; the <em>NVIC</em>", Summary: "Step through exactly what happens when an interrupt fires mid-instruction.", Part: "Talking to the World", Template: "chapter-interrupts", Interactive: true},
	{Slug: "timers-pwm", Title: "Timers &amp; <em>PWM</em>", Summary: "Move a duty-cycle slider and dim a virtual LED in real time.", Part: "Talking to the World", Template: "chapter-timers-pwm", Interactive: true},
	{Slug: "uart", Title: "UART — <em>a byte's life</em>", Summary: "Watch a single byte travel down the wire, start bit to stop bit.", Part: "Talking to the World", Template: "chapter-uart", Interactive: true},
	{Slug: "spi-i2c", Title: "SPI &amp; <em>I²C</em>", Summary: "Two ways to share a bus: dedicated lines vs addressed, acknowledged frames.", Part: "Talking to the World", Template: "chapter-spi-i2c", Interactive: true},

	{Slug: "adc", Title: "ADC — <em>sampling the analog world</em>", Summary: "Sample a smooth wave and see resolution and quantisation in action.", Part: "Going Deeper", Template: "chapter-adc", Interactive: true},
	{Slug: "low-power-volatile", Title: "Low power, <em>volatile</em> &amp; the C runtime", Summary: "Sleep modes, why volatile exists, and what runs before main().", Part: "Going Deeper", Template: "chapter-low-power-volatile"},
}

// bySlug indexes chapters for O(1) lookup. The init also stamps each
// chapter's 1-based Num from its slice position, so numbering stays correct
// no matter how chapters are reordered.
var bySlug = func() map[string]int {
	m := make(map[string]int, len(Chapters))
	for i := range Chapters {
		Chapters[i].Num = i + 1
		m[Chapters[i].Slug] = i
	}
	return m
}()

// Find returns the chapter with the given slug, plus its 1-based number, and
// whether it was found.
func Find(slug string) (Chapter, int, bool) {
	i, ok := bySlug[slug]
	if !ok {
		return Chapter{}, 0, false
	}
	return Chapters[i], i + 1, true
}

// Neighbours returns the previous and next chapters for the given slug. Either
// may be the zero Chapter (check Slug != "") at the ends of the guide.
func Neighbours(slug string) (prev, next Chapter) {
	i, ok := bySlug[slug]
	if !ok {
		return
	}
	if i > 0 {
		prev = Chapters[i-1]
	}
	if i < len(Chapters)-1 {
		next = Chapters[i+1]
	}
	return
}

// Parts returns chapters grouped into their parts, preserving order.
func Parts() []Part {
	var parts []Part
	index := map[string]int{}
	for _, c := range Chapters {
		i, ok := index[c.Part]
		if !ok {
			i = len(parts)
			index[c.Part] = i
			parts = append(parts, Part{Name: c.Part})
		}
		parts[i].Chapters = append(parts[i].Chapters, c)
	}
	return parts
}

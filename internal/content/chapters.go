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
	// LearningMinutes is an editorial estimate for reading and on-page practice;
	// optional hardware work and tool installation take additional time.
	LearningMinutes int
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
	{Slug: "what-is-embedded", Title: "What <em>is</em> an embedded system?", Summary: "The anatomy of a microcontroller, and how it differs from the computer on your desk.", LearningMinutes: 10, Part: "Foundations", Template: "chapter-what-is-embedded"},
	{Slug: "electronics-essentials", Title: "Electronics <em>essentials</em>", Summary: "Voltage, current, resistors, and the safe path from a GPIO pin to an LED.", LearningMinutes: 20, Part: "Foundations", Template: "chapter-electronics-essentials", Interactive: true},
	{Slug: "bits-and-bytes", Title: "Bits, bytes, hex &amp; <em>two's complement</em>", Summary: "How numbers really live in memory — and why negative numbers wrap around.", LearningMinutes: 15, Part: "Foundations", Template: "chapter-bits-and-bytes"},
	{Slug: "embedded-c", Title: "C for <em>embedded systems</em>", Summary: "Fixed-width types, pointers, unsigned masks, and the limits of volatile.", LearningMinutes: 25, Part: "Foundations", Template: "chapter-embedded-c", Interactive: true},
	{Slug: "bitwise-and-registers", Title: "Bitwise ops &amp; <em>registers</em>", Summary: "Set, clear and toggle individual bits on a live 32-bit hardware register.", LearningMinutes: 15, Part: "Foundations", Template: "chapter-bitwise-and-registers", Interactive: true},

	{Slug: "memory-map", Title: "The memory map &amp; <em>memory-mapped I/O</em>", Summary: "Flash vs RAM, and why writing to an address can blink an LED.", LearningMinutes: 15, Part: "The Machine", Template: "chapter-memory-map"},
	{Slug: "boot-sequence", Title: "Boot &amp; <em>startup</em>", Summary: "Step through the reset vector and the .data / .bss initialisation dance.", LearningMinutes: 15, Part: "The Machine", Template: "chapter-boot-sequence", Interactive: true},
	{Slug: "clocks", Title: "Clocks, oscillators &amp; <em>prescalers</em>", Summary: "Drag a prescaler and watch the derived peripheral frequency change.", LearningMinutes: 15, Part: "The Machine", Template: "chapter-clocks", Interactive: true},

	{Slug: "gpio", Title: "GPIO — <em>talking to a pin</em>", Summary: "Push-pull, open-drain and pull-ups, with a pin you can toggle yourself.", LearningMinutes: 15, Part: "Talking to the World", Template: "chapter-gpio", Interactive: true},
	{Slug: "interrupts", Title: "Interrupts &amp; the <em>NVIC</em>", Summary: "Step through how the core handles an interrupt while foreground code is running.", LearningMinutes: 20, Part: "Talking to the World", Template: "chapter-interrupts", Interactive: true},
	{Slug: "timers-pwm", Title: "Timers &amp; <em>PWM</em>", Summary: "Move a duty-cycle slider and dim a virtual LED in real time.", LearningMinutes: 20, Part: "Talking to the World", Template: "chapter-timers-pwm", Interactive: true},
	{Slug: "uart", Title: "UART — <em>a byte's life</em>", Summary: "Watch a single byte travel down the wire, start bit to stop bit.", LearningMinutes: 15, Part: "Talking to the World", Template: "chapter-uart", Interactive: true},
	{Slug: "spi-i2c", Title: "SPI &amp; <em>I²C</em>", Summary: "Two ways to share a bus: dedicated lines vs addressed, acknowledged frames.", LearningMinutes: 20, Part: "Talking to the World", Template: "chapter-spi-i2c", Interactive: true},

	{Slug: "adc", Title: "ADC — <em>sampling the analog world</em>", Summary: "Sample a smooth wave and see resolution and quantisation in action.", LearningMinutes: 15, Part: "Going Deeper", Template: "chapter-adc", Interactive: true},
	{Slug: "low-power-volatile", Title: "Low power, <em>volatile</em> &amp; the C runtime", Summary: "Sleep modes, why volatile exists, and what runs before main().", LearningMinutes: 20, Part: "Going Deeper", Template: "chapter-low-power-volatile"},

	{Slug: "mcu-datasheet", Title: "Cortex-M &amp; <em>reading a datasheet</em>", Summary: "Architecture, core, chip, and board—and the memory and register terminology that connects them.", LearningMinutes: 20, Part: "Practical Firmware Development", Template: "chapter-mcu-datasheet"},
	{Slug: "bare-metal-firmware", Title: "How <em>bare-metal firmware</em> works", Summary: "From reset to the main loop, with equivalent hardware-access examples in C, Rust, and Zig.", LearningMinutes: 25, Part: "Practical Firmware Development", Template: "chapter-bare-metal-firmware", Interactive: true},
	{Slug: "build-flash-debug", Title: "Build, flash &amp; <em>debug</em>", Summary: "Turn source code into a running image, then stop the chip and inspect it.", LearningMinutes: 20, Part: "Practical Firmware Development", Template: "chapter-build-flash-debug", Interactive: true},
	{Slug: "first-project", Title: "Your first <em>hardware project</em>", Summary: "Bring up a NUCLEO-F401RE, blink its LED, and move beyond blocking delays.", LearningMinutes: 25, Part: "Practical Firmware Development", Template: "chapter-first-project"},

	{Slug: "debouncing", Title: "Buttons &amp; <em>debouncing</em>", Summary: "Turn noisy switch samples into one deliberate button press without blocking the main loop.", LearningMinutes: 15, Part: "Reliable Firmware", Template: "chapter-debouncing", Interactive: true},
	{Slug: "watchdogs", Title: "Watchdogs &amp; <em>recovering from failure</em>", Summary: "Detect stalled work, choose a recovery deadline, and make resets safe and diagnosable.", LearningMinutes: 15, Part: "Reliable Firmware", Template: "chapter-watchdogs"},
	{Slug: "dma", Title: "DMA &amp; <em>buffer ownership</em>", Summary: "Move peripheral data without copying every byte on the CPU, then hand buffers over safely.", LearningMinutes: 20, Part: "Reliable Firmware", Template: "chapter-dma"},
	{Slug: "rtos", Title: "RTOS <em>fundamentals</em>", Summary: "Understand tasks, priorities, blocking, queues, and when a scheduler is worth the complexity.", LearningMinutes: 20, Part: "Reliable Firmware", Template: "chapter-rtos"},

	{Slug: "glossary", Title: "Embedded systems <em>glossary</em>", Summary: "A linked quick reference for the hardware, firmware, and concurrency terms used throughout this guide.", LearningMinutes: 15, Part: "Reference", Template: "chapter-glossary"},
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

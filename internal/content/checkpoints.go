package content

// Checkpoint is an ungraded review, separate from tracked reading sections so
// adding one does not invalidate a reader's existing section progress.
type Checkpoint struct {
	Question, Answer, Practice string
}

// Review returns the chapter-specific review shown by the shared page template.
func (c Chapter) Review() Checkpoint { return checkpoints[c.Slug] }

var checkpoints = map[string]Checkpoint{
	"what-is-embedded": {
		"How are a microcontroller, its CPU core, and a development board different?",
		"The core executes instructions. The MCU integrates that core with memory and peripherals. The board adds wiring, power, connectors, and often a debugger around the MCU.",
		"Choose a device you own and identify one input, one decision, and one physical output its firmware controls.",
	},
	"electronics-essentials": {
		"Why does an LED usually need a series resistor, even when connected to a GPIO?",
		"The resistor limits current. Size it using supply voltage, LED forward voltage, and a safe target current, then check the MCU's per-pin and total current limits. An absolute maximum is not a design target.",
		"For a 3.3 V supply, a 2.0 V LED, and 2 mA, calculate 650 ohms; choose a suitable standard value and check resistor power and GPIO limits before wiring.",
	},
	"bits-and-bytes": {
		"What does the bit pattern 0xFF mean as uint8_t versus int8_t on a two's-complement target?",
		"As an unsigned 8-bit value it is 255; as a signed two's-complement 8-bit value it is -1. Interpretation depends on width and type, not just the stored bits.",
		"Write 0xA5 in binary, then identify its high and low nibbles.",
	},
	"embedded-c": {
		"Does volatile make a shared counter increment atomic?",
		"No. It preserves required observable accesses, but a read-modify-write increment can still race. Use a suitable atomic operation or a target-appropriate synchronization mechanism.",
		"Explain why a 32-bit MMIO access needs the correct address, alignment, access width, and documented register semantics.",
	},
	"bitwise-and-registers": {
		"How do you set bit 5 without intentionally changing the other bits of an ordinary read/write value?",
		"Use value |= (1u << 5). For hardware, first check whether read-modify-write is permitted: W1C flags, side-effecting reads, and concurrent writers need different handling.",
		"Starting with 0xA5, clear bit 2, set bit 4, and predict the result before testing it.",
	},
	"memory-map": {
		"Why can storing to an address change a pin instead of storing ordinary RAM data?",
		"The memory map assigns some addresses to peripheral registers. The bus routes the access to hardware, whose register rules determine the side effect.",
		"Find your exact MCU's flash, SRAM, and GPIO address ranges in its documentation; label each source.",
	},
	"boot-sequence": {
		"Why must startup copy .data and clear .bss before ordinary C code runs?",
		"Initialized writable objects need their initial values copied into RAM, and zero-initialized objects need zeroed storage. Startup establishes those language-runtime expectations before main.",
		"Trace reset vector → reset handler → runtime initialization → main, and identify where the stack pointer is established on Cortex-M.",
	},
	"clocks": {
		"Can you assume the CPU and every timer run at the same frequency?",
		"No. Clock sources, PLLs, bus dividers, peripheral selectors, and timer multiplier rules can produce different rates. Follow the clock tree for the exact peripheral.",
		"Given a 16 MHz input and a divide-by-8 prescaler, predict 2 MHz and a 0.5 microsecond tick.",
	},
	"gpio": {
		"Why does an open-drain output need a pull-up to produce a reliable high level?",
		"It can actively pull the line low but releases it for high. A suitable pull-up supplies the high level; its resistance and bus capacitance affect rise time.",
		"Before wiring a switch, draw its pressed and released paths and check the input voltage limits and pull configuration.",
	},
	"interrupts": {
		"Why should an interrupt handler usually defer lengthy processing?",
		"Long handlers increase latency and delay other work. Capture or acknowledge the event as documented, save minimal state, and let foreground code or a task do the longer operation.",
		"Sketch a button ISR that records an event rather than delaying or printing, then identify how the foreground safely receives it.",
	},
	"timers-pwm": {
		"At a 1 MHz timer input, PSC = 0 and ARR = 999, what is the edge-aligned upcounting period?",
		"There are 1000 counter ticks per period, so the period is 1 ms and the frequency is 1 kHz. In the ideal active-high CNT < CCR model, CCR = 250 gives 25% duty.",
		"Explain why counting 0 through ARR takes ARR + 1 ticks, and why center-aligned mode needs a different calculation.",
	},
	"uart": {
		"How long does an 8N1 byte take at 9600 baud?",
		"One start bit, eight data bits, and one stop bit make ten bit times: about 1.042 ms. The theoretical payload rate is 960 bytes/s, before application gaps or protocol overhead.",
		"Encode 0x41 least-significant bit first and identify the idle, start, data, and stop levels.",
	},
	"spi-i2c": {
		"What distinguishes an I²C target address from an SPI chip-select signal?",
		"I²C carries an address on shared data/clock lines. Conventional SPI selects a target with a separate chip-select line; the SPI electrical interface itself does not define a universal address field.",
		"For a chosen sensor, record its interface voltage, clock limits, address or chip select, and required SPI mode or I²C pull-ups.",
	},
	"adc": {
		"Does increasing ADC resolution alone eliminate measurement error?",
		"No. It reduces the ideal quantization step, but reference error, noise, source impedance, acquisition time, offset, and aliasing can still dominate.",
		"Calculate the nominal step Vref / 4096 for a 12-bit ADC with a 3.3 V reference, and name two real-world error sources.",
	},
	"low-power-volatile": {
		"Does executing WFI by itself select the lowest-power mode?",
		"No. Power configuration and valid wake sources determine the mode. Retention, wake-up behavior, and clock restoration are device-specific; some modes wake through reset.",
		"Create a sleep checklist: wake source, retained state, clock recovery, pending events, and board-level current contributors.",
	},
	"mcu-datasheet": {
		"Which documents establish a pin's voltage limits, register layout, and connection to a board LED?",
		"The device datasheet gives electrical limits; the reference manual describes peripheral registers; the board schematic shows the LED connection. Check errata and board revision as well.",
		"Record your exact part, board revision, GPIO base address, output register offset, and the source for each fact.",
	},
	"bare-metal-firmware": {
		"What must exist before a register-write snippet becomes a bootable firmware image?",
		"It needs target-specific startup, a valid vector table where required, memory/linker configuration, runtime initialization, and peripheral initialization. An MMIO snippet alone is not a complete image.",
		"Trace one GPIO write from source through the linker map to the peripheral address; distinguish its code location from its destination register.",
	},
	"build-flash-debug": {
		"How are a successful build, successful flash, and running firmware different?",
		"A build creates an image; programming transfers it to the target; execution requires a valid boot path and a running core. A breakpoint, wrong image, or startup fault can prevent visible behavior after programming succeeds.",
		"Set a breakpoint in main, confirm the programmed ELF, inspect a variable, then resume and test a physical reset.",
	},
	"first-project": {
		"Why does toggling an LED every 500 ms produce a one-second cycle?",
		"A full cycle has two transitions: off to on and on to off. Unsigned elapsed-time checks let the loop do other short work while waiting for the next transition.",
		"Change the heartbeat interval, rebuild and program, then verify the deliberate change survives a physical reset without an attached debug session.",
	},
	"debouncing": {
		"Why can one physical press cause several GPIO edge interrupts?",
		"Mechanical contacts can bounce between levels before settling. Treat edges as observations, not accepted presses; require a stable interval or a defined run of consistent samples before changing the debounced state.",
		"Step through the sample explorer. Count raw rising edges and accepted presses, then test a pulse shorter than the acceptance threshold.",
	},
	"watchdogs": {
		"Why is feeding a watchdog unconditionally from a timer ISR risky?",
		"The timer may keep running while the application is stuck. Refresh should depend on evidence that critical work completed within the required health window, not merely that an interrupt still executes.",
		"Define two health milestones, a recovery deadline, and a safe reset state; describe how you would inject a fault to verify them.",
	},
	"dma": {
		"When may software reuse a buffer submitted to DMA?",
		"Only after the relevant documented transfer or handoff guarantees DMA is no longer using that region. Also account for peripheral completion, memory ordering, and cache maintenance on targets that require it.",
		"Draw a two-buffer ownership timeline and decide what happens when the CPU has not finished a buffer before DMA needs it again.",
	},
	"rtos": {
		"Why does a high-priority task that never blocks risk starving other tasks?",
		"A priority-based preemptive scheduler chooses the highest-priority ready task. Lower-priority tasks may never run if that task stays ready; equal-priority time slicing does not solve this.",
		"Split a sensor → queue → logger pipeline into tasks. Assign ownership, queue capacity, an overflow policy, and maximum blocking times before assigning priorities.",
	},
	"glossary": {
		"How are volatile, atomicity, and synchronization different?",
		"Volatile concerns observable accesses. Atomicity concerns whether an operation is indivisible to relevant observers. Synchronization coordinates access and ordering between execution contexts. None is a substitute for all the others.",
		"Choose five unfamiliar terms, explain each without looking, and follow its lesson link to check your explanation.",
	},
}

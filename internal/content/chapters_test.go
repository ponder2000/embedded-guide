package content

import "testing"

func TestNumbersAreSequential(t *testing.T) {
	for i, c := range Chapters {
		if c.Num != i+1 {
			t.Errorf("chapter %q: Num = %d, want %d", c.Slug, c.Num, i+1)
		}
	}
}

func TestSlugsUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range Chapters {
		if seen[c.Slug] {
			t.Errorf("duplicate slug: %q", c.Slug)
		}
		seen[c.Slug] = true
	}
}

func TestFind(t *testing.T) {
	first := Chapters[0]
	got, num, ok := Find(first.Slug)
	if !ok || got.Slug != first.Slug || num != 1 {
		t.Errorf("Find(%q) = %+v, %d, %v", first.Slug, got, num, ok)
	}
	if _, _, ok := Find("does-not-exist"); ok {
		t.Error("Find returned ok for unknown slug")
	}
}

func TestNeighbours(t *testing.T) {
	// First chapter has no prev; last has no next.
	if prev, _ := Neighbours(Chapters[0].Slug); prev.Slug != "" {
		t.Errorf("first chapter prev should be empty, got %q", prev.Slug)
	}
	if _, next := Neighbours(Chapters[len(Chapters)-1].Slug); next.Slug != "" {
		t.Errorf("last chapter next should be empty, got %q", next.Slug)
	}
	// A middle chapter links to its neighbours by slice order.
	prev, next := Neighbours(Chapters[1].Slug)
	if prev.Slug != Chapters[0].Slug || next.Slug != Chapters[2].Slug {
		t.Errorf("middle neighbours wrong: prev=%q next=%q", prev.Slug, next.Slug)
	}
}

func TestPartsCoverAllChapters(t *testing.T) {
	var total int
	for _, p := range Parts() {
		if p.Name == "" {
			t.Error("part with empty name")
		}
		total += len(p.Chapters)
	}
	if total != len(Chapters) {
		t.Errorf("parts cover %d chapters, want %d", total, len(Chapters))
	}
}

func TestBeginnerLearningPath(t *testing.T) {
	for _, slug := range []string{"electronics-essentials", "embedded-c", "mcu-datasheet", "bare-metal-firmware", "build-flash-debug", "first-project"} {
		chapter, _, ok := Find(slug)
		if !ok || chapter.Summary == "" || chapter.Template != "chapter-"+slug {
			t.Errorf("beginner lesson %q is missing or incomplete", slug)
		}
	}
	for _, pair := range [][2]string{{"what-is-embedded", "electronics-essentials"}, {"electronics-essentials", "bitwise-and-registers"}, {"embedded-c", "memory-map"}, {"mcu-datasheet", "bare-metal-firmware"}, {"bare-metal-firmware", "build-flash-debug"}, {"build-flash-debug", "first-project"}} {
		_, before, _ := Find(pair[0])
		_, after, _ := Find(pair[1])
		if before >= after {
			t.Errorf("%s should come before %s", pair[0], pair[1])
		}
	}
	for _, slug := range []string{"mcu-datasheet", "bare-metal-firmware", "build-flash-debug", "first-project"} {
		chapter, _, _ := Find(slug)
		if chapter.Part != "Practical Firmware Development" {
			t.Errorf("%s should be grouped under Practical Firmware Development", slug)
		}
	}
}

func TestEveryChapterHasReview(t *testing.T) {
	for _, chapter := range Chapters {
		if chapter.LearningMinutes <= 0 {
			t.Errorf("%s needs a positive learning-time estimate", chapter.Slug)
		}
		review := chapter.Review()
		if review.Question == "" || review.Answer == "" || review.Practice == "" {
			t.Errorf("%s is missing its question, explanation, or practice", chapter.Slug)
		}
	}
	for slug := range checkpoints {
		if _, _, ok := Find(slug); !ok {
			t.Errorf("review belongs to unregistered chapter %s", slug)
		}
	}
}

func TestReliableFirmwarePath(t *testing.T) {
	_, previous, _ := Find("first-project")
	for _, slug := range []string{"debouncing", "watchdogs", "dma", "rtos", "glossary"} {
		chapter, number, ok := Find(slug)
		if !ok || number <= previous || chapter.Template != "chapter-"+slug {
			t.Errorf("missing or incorrectly ordered follow-up lesson %s", slug)
		}
		if slug != "glossary" && chapter.Part != "Reliable Firmware" {
			t.Errorf("%s should be in Reliable Firmware", slug)
		}
		previous = number
	}
}

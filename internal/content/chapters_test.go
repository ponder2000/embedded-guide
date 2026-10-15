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

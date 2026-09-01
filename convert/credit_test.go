package convert

import "testing"

func TestImageCreditFromMatchedFigure(t *testing.T) {
	html := `<figure><img src="lead.jpg"><figcaption>Photo: Jane Doe</figcaption></figure>` +
		`<figure><img src="other.jpg"><figcaption>Not this one</figcaption></figure>`
	if got := ImageCredit(html, "lead.jpg"); got != "Photo: Jane Doe" {
		t.Errorf("ImageCredit = %q, want the matched figure's caption", got)
	}
}

func TestImageCreditPrefersMatchedOverFirst(t *testing.T) {
	html := `<figure><img src="a.jpg"><figcaption>First</figcaption></figure>` +
		`<figure><img src="lead.jpg"><figcaption>Second</figcaption></figure>`
	if got := ImageCredit(html, "lead.jpg"); got != "Second" {
		t.Errorf("ImageCredit = %q, want the caption of the matching figure", got)
	}
}

func TestImageCreditDoesNotBorrowAnotherFigure(t *testing.T) {
	html := `<figure><img src="a.jpg"><figcaption>First</figcaption></figure>` +
		`<figure><img src="b.jpg"><figcaption>Second</figcaption></figure>`
	// lead.jpg is in no figure: its caption must be empty rather than the
	// first figure's, so unrelated images never display another image's
	// caption.
	if got := ImageCredit(html, "lead.jpg"); got != "" {
		t.Errorf("ImageCredit = %q, want empty (no figure match, no borrowing)", got)
	}
}

func TestImageCreditIgnoresCreditClassOutsideFigure(t *testing.T) {
	html := `<p>body</p><div class="image-credit">AP Photo</div>`
	// The credit-class element is not an image's own credit; a document-global
	// probe would attach it to every image.
	if got := ImageCredit(html, ""); got != "" {
		t.Errorf("ImageCredit = %q, want empty (no matched figure)", got)
	}
}

func TestImageCreditCollapsesWhitespace(t *testing.T) {
	html := `<figure><img src="lead.jpg"><figcaption>
		Photo:
		Jane Doe
	</figcaption></figure>`
	if got := ImageCredit(html, "lead.jpg"); got != "Photo: Jane Doe" {
		t.Errorf("ImageCredit = %q, want collapsed single line", got)
	}
}

func TestImageCaptionPhotoGridAttribution(t *testing.T) {
	// A photo-grid: one figure wrapping two photos with a single combined
	// caption. The caption belongs to the last image; the first has none.
	html := `<figure><img src="left.jpg"><img src="right.jpg"><figcaption>Left: one. Right: two.</figcaption></figure>`
	if got := ImageCredit(html, "left.jpg"); got != "" {
		t.Errorf("ImageCredit(left.jpg) = %q, want empty (caption belongs to the last image)", got)
	}
	if got := ImageCredit(html, "right.jpg"); got != "Left: one. Right: two." {
		t.Errorf("ImageCredit(right.jpg) = %q, want the figure's caption", got)
	}
}

func TestImageCaptionNestedPhotoGrid(t *testing.T) {
	// The real Intercept photo-grid structure: two inner per-photo figures
	// wrapped by an outer figure carrying the combined caption. The caption
	// belongs to the second (last) photo; the first renders with none.
	html := `<figure class="photo-grid">` +
		`<figure><img src="2024.jpg"></figure>` +
		`<figure><img src="2026.jpg"></figure>` +
		`<figcaption>Left: 2024. Right: 2026.</figcaption>` +
		`</figure>`
	if got := ImageCredit(html, "2024.jpg"); got != "" {
		t.Errorf("ImageCredit(2024.jpg) = %q, want empty", got)
	}
	if got := ImageCredit(html, "2026.jpg"); got != "Left: 2024. Right: 2026." {
		t.Errorf("ImageCredit(2026.jpg) = %q, want the combined caption", got)
	}
}

func TestImageCaptionReportsSharedFigureNonOwner(t *testing.T) {
	// The first image of a grid is inside a captioned figure but is not the
	// caption owner: ImageCaption reports matched with an empty caption so the
	// caller renders no caption instead of falling back to the source label.
	html := `<figure><img src="left.jpg"><img src="right.jpg"><figcaption>Cap</figcaption></figure>`
	if cap, matched := ImageCaption(html, "left.jpg"); !matched || cap != "" {
		t.Errorf("ImageCaption(left.jpg) = (%q, %v), want (\"\", true)", cap, matched)
	}
	if cap, matched := ImageCaption(html, "right.jpg"); !matched || cap != "Cap" {
		t.Errorf("ImageCaption(right.jpg) = (%q, %v), want (\"Cap\", true)", cap, matched)
	}
	if cap, matched := ImageCaption(html, "other.jpg"); matched || cap != "" {
		t.Errorf("ImageCaption(other.jpg) = (%q, %v), want (\"\", false)", cap, matched)
	}
}

func TestImageCreditEmptyWhenNoCredit(t *testing.T) {
	cases := []string{
		`<p>plain body</p>`,
		`<figure><img src="lead.jpg"></figure>`,
		`<div class="caption">not a credit class</div>`,
	}
	for _, html := range cases {
		if got := ImageCredit(html, "lead.jpg"); got != "" {
			t.Errorf("ImageCredit(%q) = %q, want empty", html, got)
		}
	}
}

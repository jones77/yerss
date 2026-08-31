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

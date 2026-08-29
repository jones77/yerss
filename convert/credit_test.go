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

func TestImageCreditFallsBackToFirstFigure(t *testing.T) {
	html := `<figure><img src="a.jpg"><figcaption>First</figcaption></figure>` +
		`<figure><img src="b.jpg"><figcaption>Second</figcaption></figure>`
	if got := ImageCredit(html, "lead.jpg"); got != "First" {
		t.Errorf("ImageCredit = %q, want the first figure's caption", got)
	}
}

func TestImageCreditFromCreditClass(t *testing.T) {
	html := `<p>body</p><div class="image-credit">AP Photo</div>`
	if got := ImageCredit(html, ""); got != "AP Photo" {
		t.Errorf("ImageCredit = %q, want the credit-class text", got)
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

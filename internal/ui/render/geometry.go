package render

// ContentGeom computes the article content-area geometry shared by the article
// state builder, the mouse coordinate mapping, and the border renderer. textW
// is the content column width inside the horizontal padding, viewportH the
// number of content rows (the full interior between the top and bottom borders;
// the article view has no vertical padding). Keeping the computation in one
// place guarantees mouse hit-testing and rendering agree.
func ContentGeom(w, h, padX int) (textW, viewportH int) {
	textW = max(1, w-2*padX-2)
	viewportH = max(1, h-2)
	return textW, viewportH
}

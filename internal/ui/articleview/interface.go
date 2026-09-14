package articleview

import "yerss/internal/ui/imageview"

// State implements imageview.ArticleView so the image machinery can read the
// composed article (blocks, anchors, viewport) without depending on this
// package. The interface is defined on the imageview side to avoid an import
// cycle, since articleview imports imageview for block composition.
var _ imageview.ArticleView = (*State)(nil)
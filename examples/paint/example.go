// Package paint provides examples of how to use the matcha/paint package.
package paint

import (
	"golang.org/x/image/colornames"
	"gomatcha.io/bridge"
	"gomatcha.io/matcha/layout"
	"gomatcha.io/matcha/layout/constraint"
	"gomatcha.io/matcha/paint"
	"gomatcha.io/matcha/view"
	"gomatcha.io/matcha/view/basicview"
)

func init() {
	bridge.RegisterFunc("gomatcha.io/matcha/examples/paint New", func() *view.Root {
		return view.NewRoot(New())
	})
}

type PaintView struct {
	view.Embed
}

func New() *PaintView {

	return &PaintView{}
}

func (v *PaintView) Build(ctx *view.Context) view.Model {
	l := &constraint.Layouter{}

	chl1 := basicview.New()
	chl1.Painter = &paint.Style{
		Transparency:    0.1,
		BackgroundColor: colornames.Blue,
		BorderColor:     colornames.Red,
		BorderWidth:     3,
		CornerRadius:    20,
		ShadowRadius:    4,
		ShadowOffset:    layout.Pt(5, 5),
		ShadowColor:     colornames.Black,
	}
	g1 := l.Add(chl1, func(s *constraint.Solver) {
		s.TopEqual(constraint.Const(100))
		s.LeftEqual(constraint.Const(100))
		s.WidthEqual(constraint.Const(100))
		s.HeightEqual(constraint.Const(100))
	})

	chl2 := basicview.New()
	chl2.Painter = &paint.Style{BackgroundColor: colornames.Yellow}
	g2 := l.Add(chl2, func(s *constraint.Solver) {
		s.TopEqual(g1.Bottom())
		s.LeftEqual(g1.Left())
		s.WidthEqual(constraint.Const(100))
		s.HeightEqual(constraint.Const(100))
	})

	chl3 := basicview.New()
	chl3.Painter = &paint.Style{BackgroundColor: colornames.Blue}
	g3 := l.Add(chl3, func(s *constraint.Solver) {
		s.TopEqual(g2.Bottom())
		s.LeftEqual(g2.Left())
		s.WidthEqual(constraint.Const(100))
		s.HeightEqual(constraint.Const(100))
	})

	chl4 := basicview.New()
	chl4.Painter = &paint.Style{BackgroundColor: colornames.Magenta}
	_ = l.Add(chl4, func(s *constraint.Solver) {
		s.TopEqual(g2.Bottom())
		s.LeftEqual(g3.Left())
		s.WidthEqual(constraint.Const(100))
		s.HeightEqual(constraint.Const(100))
	})

	return view.Model{
		Children: l.Views(),
		Layouter: l,
		Painter:  &paint.Style{BackgroundColor: colornames.Green},
	}
}

package view

import (
	"golang.org/x/image/colornames"
	"gomatcha.io/bridge"
	"gomatcha.io/matcha/comm"
	"gomatcha.io/matcha/layout/constraint"
	"gomatcha.io/matcha/paint"
	"gomatcha.io/matcha/view"
	"gomatcha.io/matcha/view/progressview"
	"gomatcha.io/matcha/view/slider"
)

func init() {
	bridge.RegisterFunc("gomatcha.io/matcha/examples/view NewProgressView", func() *view.Root {
		return view.NewRoot(NewProgressView())
	})
}

type ProgressView struct {
	view.Embed
	value *comm.Float64Value
}

func NewProgressView() *ProgressView {
	return &ProgressView{
		value: &comm.Float64Value{},
	}
}

func (v *ProgressView) Build(ctx *view.Context) view.Model {
	l := &constraint.Layouter{}

	progressv := progressview.New()
	progressv.ProgressNotifier = v.value
	progressv.ProgressColor = colornames.Red
	l.Add(progressv, func(s *constraint.Solver) {
		s.Top(100)
		s.Left(100)
		s.Width(200)
	})

	sliderv := slider.New()
	sliderv.MaxValue = 1
	sliderv.MinValue = 0
	sliderv.OnValueChange = func(value float64) {
		v.value.SetValue(value)
	}
	l.Add(sliderv, func(s *constraint.Solver) {
		s.Top(200)
		s.Left(100)
		s.Width(200)
	})

	return view.Model{
		Children: l.Views(),
		Layouter: l,
		Painter:  &paint.Style{BackgroundColor: colornames.White},
	}
}

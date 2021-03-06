// Package todo provides an example of a basic Todo app.
package todo

import (
	"image/color"

	"golang.org/x/image/colornames"

	"gomatcha.io/bridge"
	"gomatcha.io/matcha/app"
	"gomatcha.io/matcha/keyboard"
	"gomatcha.io/matcha/layout/constraint"
	"gomatcha.io/matcha/layout/table"
	"gomatcha.io/matcha/paint"
	"gomatcha.io/matcha/text"
	"gomatcha.io/matcha/touch"
	"gomatcha.io/matcha/view"
	"gomatcha.io/matcha/view/basicview"
	"gomatcha.io/matcha/view/imageview"
	"gomatcha.io/matcha/view/scrollview"
	"gomatcha.io/matcha/view/stackview"
	"gomatcha.io/matcha/view/textinput"
	"gomatcha.io/matcha/view/textview"
)

func init() {
	bridge.RegisterFunc("gomatcha.io/matcha/examples/todo New", func() *view.Root {
		appview := NewAppView()

		v := stackview.New()
		v.Stack = &stackview.Stack{}
		v.Stack.SetViews(appview)
		v.BarColor = color.RGBA{R: 46, G: 124, B: 190, A: 1}
		v.TitleTextStyle = &text.Style{}
		v.TitleTextStyle.SetFont(text.Font{
			Family: "Helvetica Neue",
			Face:   "Medium",
			Size:   20,
		})
		v.TitleTextStyle.SetTextColor(colornames.White)
		return view.NewRoot(v)
	})
}

type Todo struct {
	Title     string
	Completed bool
}

type AppView struct {
	view.Embed
	Todos []*Todo
}

func NewAppView() *AppView {
	return &AppView{}
}

func (v *AppView) Build(ctx *view.Context) view.Model {
	l := &table.Layouter{}

	for i, todo := range v.Todos {
		idx := i
		todoView := NewTodoView()
		todoView.Todo = todo
		todoView.OnDelete = func() {
			v.Todos = append(v.Todos[:idx], v.Todos[idx+1:]...)
			v.Signal()
		}
		todoView.OnComplete = func(complete bool) {
			v.Todos[idx].Completed = complete
			v.Signal()
		}
		l.Add(todoView, nil)
	}

	addView := NewAddView()
	addView.OnAdd = func(title string) {
		v.Todos = append(v.Todos, &Todo{Title: title})
		v.Signal()
	}
	l.Add(addView, nil)

	scrollView := scrollview.New()
	scrollView.ContentChildren = l.Views()
	scrollView.ContentLayouter = l
	return view.Model{
		Children: []view.View{scrollView},
		Painter:  &paint.Style{BackgroundColor: colornames.White},
		Options: []view.Option{
			app.StatusBar{Style: app.StatusBarStyleLight},
		},
	}
}

func (v *AppView) StackBar(ctx *view.Context) *stackview.Bar {
	return &stackview.Bar{Title: "To Do Example"}
}

type AddView struct {
	view.Embed
	text      *text.Text
	responder keyboard.Responder
	OnAdd     func(title string)
}

func NewAddView() *AddView {
	return &AddView{
		text: text.New(""),
	}
}

func (v *AddView) Build(ctx *view.Context) view.Model {
	l := &constraint.Layouter{}
	l.Solve(func(s *constraint.Solver) {
		s.Height(50)
		s.WidthEqual(l.MaxGuide().Width())
	})

	style := &text.Style{}
	style.SetFont(text.Font{
		Family: "Helvetica Neue",
		Size:   20,
	})

	placeholderStyle := &text.Style{}
	placeholderStyle.SetFont(text.Font{
		Family: "Helvetica Neue",
		Size:   20,
	})
	placeholderStyle.SetTextColor(colornames.Lightgray)

	input := textinput.New()
	input.PaintStyle = &paint.Style{BackgroundColor: colornames.White}
	input.Text = v.text
	input.Style = style
	input.PlaceholderText = text.New("What needs to be done?")
	input.PlaceholderStyle = placeholderStyle
	input.KeyboardReturnType = keyboard.DoneReturnType
	input.Responder = &v.responder
	input.OnSubmit = func() {
		str := v.text.String()
		v.responder.Dismiss()
		v.text.SetString("")
		if str != "" {
			v.OnAdd(str)
		}
	}
	l.Add(input, func(s *constraint.Solver) {
		s.LeftEqual(l.Left().Add(15))
		s.RightEqual(l.Right().Add(-15))
		s.CenterYEqual(l.CenterY())
	})

	separator := basicview.New()
	separator.Painter = &paint.Style{BackgroundColor: color.RGBA{203, 202, 207, 255}}
	l.Add(separator, func(s *constraint.Solver) {
		s.Height(1)
		s.LeftEqual(l.Left())
		s.RightEqual(l.Right())
		s.BottomEqual(l.Bottom())
	})

	return view.Model{
		Children: l.Views(),
		Layouter: l,
	}
}

type TodoView struct {
	view.Embed
	Todo       *Todo
	OnDelete   func()
	OnComplete func(check bool)
}

func NewTodoView() *TodoView {
	return &TodoView{}
}

func (v *TodoView) Build(ctx *view.Context) view.Model {
	l := &constraint.Layouter{}
	l.Solve(func(s *constraint.Solver) {
		s.Height(50)
		s.WidthEqual(l.MaxGuide().Width())
	})

	checkbox := NewCheckbox()
	checkbox.Value = v.Todo.Completed
	checkbox.OnValueChange = func(value bool) {
		v.OnComplete(value)
	}
	checkboxGuide := l.Add(checkbox, func(s *constraint.Solver) {
		s.CenterYEqual(l.CenterY())
		s.LeftEqual(l.Left().Add(15))
	})

	deleteButton := NewDeleteButton()
	deleteButton.OnPress = func() {
		v.OnDelete()
	}
	deleteGuide := l.Add(deleteButton, func(s *constraint.Solver) {
		s.CenterYEqual(l.CenterY())
		s.RightEqual(l.Right().Add(-15))
	})

	titleView := textview.New()
	titleView.String = v.Todo.Title
	titleView.Style = nil //...
	l.Add(titleView, func(s *constraint.Solver) {
		s.CenterYEqual(l.CenterY())
		s.LeftEqual(checkboxGuide.Right().Add(15))
		s.RightEqual(deleteGuide.Left().Add(-15))
	})

	separator := basicview.New()
	separator.Painter = &paint.Style{BackgroundColor: color.RGBA{203, 202, 207, 255}}
	l.Add(separator, func(s *constraint.Solver) {
		s.Height(1)
		s.LeftEqual(l.Left())
		s.RightEqual(l.Right())
		s.BottomEqual(l.Bottom())
	})

	return view.Model{
		Children: l.Views(),
		Layouter: l,
	}
}

type Checkbox struct {
	view.Embed
	Value         bool
	OnValueChange func(value bool)
}

func NewCheckbox() *Checkbox {
	return &Checkbox{}
}

func (v *Checkbox) Build(ctx *view.Context) view.Model {
	l := &constraint.Layouter{}
	l.Solve(func(s *constraint.Solver) {
		s.Width(40)
		s.Height(40)
	})

	imageView := imageview.New()
	if v.Value {
		imageView.Image = app.MustLoadImage("CheckboxChecked")
	} else {
		imageView.Image = app.MustLoadImage("CheckboxUnchecked")
	}
	l.Add(imageView, func(s *constraint.Solver) {
		s.CenterXEqual(l.CenterX())
		s.CenterYEqual(l.CenterY())
		s.WidthEqual(l.Width())
		s.HeightEqual(l.Height())
	})

	button := &touch.ButtonRecognizer{
		OnTouch: func(e *touch.ButtonEvent) {
			if e.Kind == touch.EventKindRecognized {
				v.OnValueChange(!v.Value)
			}
		},
	}

	return view.Model{
		Children: l.Views(),
		// Painter:  painter,
		Layouter: l,
		Options: []view.Option{
			touch.RecognizerList{button},
		},
	}
}

type DeleteButton struct {
	view.Embed
	OnPress func()
}

func NewDeleteButton() *DeleteButton {
	return &DeleteButton{}
}

func (v *DeleteButton) Build(ctx *view.Context) view.Model {
	l := &constraint.Layouter{}
	l.Solve(func(s *constraint.Solver) {
		s.Width(40)
		s.Height(40)
	})

	imageView := imageview.New()
	imageView.Image = app.MustLoadImage("Delete")
	l.Add(imageView, func(s *constraint.Solver) {
		s.CenterXEqual(l.CenterX())
		s.CenterYEqual(l.CenterY())
		s.WidthEqual(l.Width())
		s.HeightEqual(l.Height())
	})

	button := &touch.ButtonRecognizer{
		OnTouch: func(e *touch.ButtonEvent) {
			if e.Kind == touch.EventKindRecognized {
				v.OnPress()
			}
		},
	}

	return view.Model{
		Children: l.Views(),
		Layouter: l,
		Options: []view.Option{
			touch.RecognizerList{button},
		},
	}
}

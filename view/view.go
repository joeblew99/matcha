/*
Package view provides the component library.

View

View is the core building block of Matcha. It defines components
that come together to create a hierarchy of views that becomes your app. The View
interface has the following five methods.

	Build(*Context) Model

Build is the most important method. It is similar to React's render() function.
When your view updates and needs to be displayed, Build is called to get the view's children, layout,
paint style, options, etc. These are returned in the Model struct. Unlike iOS where
properties can be changed independently (label.title = @"Foo"), when a
Matcha view updates, Build() is called again and all its properties and children are
completely recreated.

	ViewKey() interface{}

ViewKey returns a view's stable identifier. This should not change over the lifetime
of the view. This allows Matcha to to track which items have been added or removed.

	Lifecycle(from, to Stage)

Lifecycle gets called as a view gets displayed or hidden. A view may cross through multiple
lifecycle stages at the same time. For example a view can start at StageDead and
jump directly to StageVisible. If the view needs to perform an action on mount,
EntersStage(from, to, StageMounted) can be used to track this transition.

	Notify(f func()) comm.Id
	Unnotify(id comm.Id)

Finally we have Notify and Unnotify which provide a way for views to signal to the framework that they
have changed and need updating. See the comm.Notifier docs for more information.

As a guideline the following is recommended.

	* Views should not keep references to their children or modify/create them outside of Build().
	* Views should not keep references to their parents.
	* Views should only be modified while the matcha.MainLocker() mutex is held.

Embed

Implementing all the View methods for every component would be a hassle, so instead we can
use Go's embedding functionality with the Embed struct to provide a basic
implementation of these methods. Embed additionally adds the Subscribe(), Unsubscribe(),
and Signal() methods to simplify signaling for updates. We see an example of this below.

	type ExampleView struct {
		view.Embed
		notifier comm.Notifier
	}
	func New(n comm.Notifier) *ExampleView {
		return &ExampleView{
			Embed: view.NewEmbed(n),
			notifier: n,
		}
	}
	func (v *ExampleView) Lifecycle(from, to view.Stage) {
		if view.EntersStage(from, to, view.StageMounted) {
			// Update anytime v.n changes.
			v.Subscribe(v.notifier)
		} else if view.ExitsStage(from, to, view.StageMounted) {
			// We must unsubscribe when the view is unmounted or risk a leak.
			v.Unsubscribe(v.notifier)
		}
	}
	func (v *ExampleView) Build(ctx *view.Context) view.Model {
		child := button.New()
		child.String = "Click me"
		child.OnClick = func() {
			// Trigger the view to rebuild when the button is clicked.
			v.Signal()
		}
		return view.Model{
			Children: []view.View{child},
		}
	}

*/
package view

import (
	"sync"

	"github.com/gogo/protobuf/proto"
	"gomatcha.io/matcha/comm"
	"gomatcha.io/matcha/layout"
	"gomatcha.io/matcha/paint"
)

type Id int64

type View interface {
	Build(*Context) Model
	Lifecycle(from, to Stage)
	ViewKey() interface{}
	comm.Notifier
}

type Option interface {
	OptionKey() string
}

// Embed is a convenience struct that provides a default implementation of View. It also wraps a comm.Relay.
type Embed struct {
	key   interface{}
	Key   interface{}
	mu    sync.Mutex
	relay comm.Relay
}

func NewEmbed(key interface{}) Embed {
	return Embed{
		key: key,
	}
}

// Build is an empty implementation of View's Build method.
func (e *Embed) Build(ctx *Context) Model {
	return Model{}
}

func (e *Embed) ViewKey() interface{} {
	return struct {
		A interface{}
		B interface{}
	}{A: e.key, B: e.Key}
}

// Lifecycle is an empty implementation of View's Lifecycle method.
func (e *Embed) Lifecycle(from, to Stage) {
	// no-op
}

// Notify calls Notify(id) on the underlying comm.Relay.
func (e *Embed) Notify(f func()) comm.Id {
	return e.relay.Notify(f)
}

// Unnotify calls Unnotify(id) on the underlying comm.Relay.
func (e *Embed) Unnotify(id comm.Id) {
	e.relay.Unnotify(id)
}

// Subscribe calls Subscribe(n) on the underlying comm.Relay.
func (e *Embed) Subscribe(n comm.Notifier) {
	e.relay.Subscribe(n)
}

// Unsubscribe calls Unsubscribe(n) on the underlying comm.Relay.
func (e *Embed) Unsubscribe(n comm.Notifier) {
	e.relay.Unsubscribe(n)
}

// Update calls Signal() on the underlying comm.Relay.
func (e *Embed) Signal() {
	e.relay.Signal()
}

type Stage int

const (
	// StageDead marks views that are not attached to the view hierarchy.
	StageDead Stage = iota
	// StageMounted marks views that are in the view hierarchy but not visible.
	StageMounted
	// StageVisible marks views that are in the view hierarchy and visible.
	StageVisible
)

// EntersStage returns true if from<s and to>=s.
func EntersStage(from, to, s Stage) bool {
	return from < s && to >= s
}

// ExitsStage returns true if from>=s and to<s.
func ExitsStage(from, to, s Stage) bool {
	return from >= s && to < s
}

// Model describes the view and its children.
type Model struct {
	Children []View
	Layouter layout.Layouter
	Painter  paint.Painter
	Options  []Option

	NativeViewName  string
	NativeViewState proto.Message
	NativeValues    map[string]proto.Message
	NativeFuncs     map[string]interface{}
}

// WithPainter wraps the view v, and replaces its Model.Painter with p.
func WithPainter(v View, p paint.Painter) View {
	return &painterView{View: v, painter: p}
}

type painterView struct {
	View
	painter paint.Painter
}

func (v *painterView) Build(ctx *Context) Model {
	m := v.View.Build(ctx)
	m.Painter = v.painter
	return m
}

// WithOptions wraps the view v, and adds the given options to its Model.Options.
func WithOptions(v View, opts []Option) View {
	return &optionsView{View: v, options: opts}
}

type optionsView struct {
	View
	options []Option
}

func (v *optionsView) Build(ctx *Context) Model {
	m := v.View.Build(ctx)
	m.Options = append(m.Options, v.options...)
	return m
}

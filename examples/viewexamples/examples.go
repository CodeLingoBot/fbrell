// Package viewexamples provides the various views on Rell examples.
package viewexamples

import (
	"errors"
	"fmt"
	"github.com/nshah/go.fburl"
	"github.com/nshah/go.h"
	"github.com/nshah/go.h.js.fb"
	"github.com/nshah/go.h.js.loader"
	"github.com/nshah/rell/context"
	"github.com/nshah/rell/examples"
	"github.com/nshah/rell/js"
	"github.com/nshah/rell/view"
	"net/http"
)

var envOptions = map[string]string{
	"Production with CDN":    "",
	"Production without CDN": fburl.Production,
	"Beta":                   fburl.Beta,
	"Latest":                 "latest",
	"Dev":                    "dev",
	"Intern":                 "intern",
	"In Your":                "inyour",
}

// Parse the Context and an Example.
func parse(r *http.Request) (*context.Context, *examples.Example, error) {
	context, err := context.FromRequest(r)
	if err != nil {
		return nil, nil, err
	}
	example, err := examples.Load(context.Version, r.URL.Path)
	if err != nil {
		return nil, nil, err
	}
	return context, example, nil
}

func List(w http.ResponseWriter, r *http.Request) {
	context, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	view.Write(w, r, renderList(context, examples.GetDB(context.Version)))
}

func Saved(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && r.URL.Path == "/saved/" {
		c, err := context.FromRequest(r)
		if err != nil {
			view.Error(w, r, err)
			return
		}
		hash, err := examples.Save([]byte(r.FormValue("code")))
		if err != nil {
			view.Error(w, r, err)
			return
		}
		http.Redirect(w, r, c.ViewURL("/saved/"+hash), 302)
		return
	} else {
		context, example, err := parse(r)
		if err != nil {
			view.Error(w, r, err)
			return
		}
		view.Write(w, r, renderExample(context, example))
	}
}

func Raw(w http.ResponseWriter, r *http.Request) {
	_, example, err := parse(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	if !example.AutoRun {
		view.Error(
			w, r, errors.New("Not allowed to view this example in raw mode."))
		return
	}
	w.Write(example.Content)
}

func Simple(w http.ResponseWriter, r *http.Request) {
	context, example, err := parse(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	if !example.AutoRun {
		view.Error(
			w, r, errors.New("Not allowed to view this example in simple mode."))
		return
	}
	view.Write(w, r, &h.Document{
		Inner: &h.Frag{
			&h.Head{
				Inner: &h.Frag{
					&h.Meta{Charset: "utf-8"},
					&h.Title{h.String(example.Title)},
				},
			},
			&h.Body{
				Inner: &h.Frag{
					&loader.HTML{
						Resource: []loader.Resource{
							&fb.Init{
								AppID:      context.AppID,
								ChannelURL: context.ChannelURL(),
								URL:        context.SdkURL(),
							},
						},
					},
					&h.Div{
						ID:    "example",
						Inner: h.UnsafeBytes(example.Content),
					},
				},
			},
		},
	})
}

func SdkChannel(w http.ResponseWriter, r *http.Request) {
	const maxAge = 31536000 // 1 year
	context, err := context.FromRequest(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	view.Write(w, r, &h.Script{Src: context.SdkURL()})
}

func Example(w http.ResponseWriter, r *http.Request) {
	context, example, err := parse(r)
	if err != nil {
		view.Error(w, r, err)
		return
	}
	view.Write(w, r, renderExample(context, example))
}

func renderExample(c *context.Context, example *examples.Example) *view.Page {
	return &view.Page{
		Title:    example.Title,
		Class:    "main",
		Resource: []loader.Resource{&js.Init{Context: c, Example: example}},
		Body: &h.Frag{
			&h.Div{
				ID: "interactive",
				Inner: &h.Frag{
					renderEnvSelector(c, example),
					&h.Div{
						Class: "controls",
						Inner: &h.Frag{
							&h.A{ID: "rell-login", Inner: &h.Span{Inner: h.String("Login")}},
							&h.Span{ID: "auth-status-label", Inner: h.String("Status:")},
							&h.Span{ID: "auth-status", Inner: h.String("waiting")},
							&h.Span{Class: "bar", Inner: h.String("|")},
							&h.A{
								ID:    "rell-disconnect",
								Inner: h.String("Disconnect"),
							},
							&h.Span{Class: "bar", Inner: h.String("|")},
							&h.A{
								ID:    "rell-logout",
								Inner: h.String("Logout"),
							},
						},
					},
					&h.Form{
						Action: c.URL("/saved/").String(),
						Method: h.Post,
						Target: "_top",
						Inner: &h.Frag{
							h.HiddenInputs(c.Values()),
							&h.Textarea{
								ID:    "jscode",
								Name:  "code",
								Inner: h.String(example.Content),
							},
							&h.Div{
								Class: "controls",
								Inner: &h.Frag{
									&h.Strong{
										Inner: &h.A{
											HREF:  c.URL("/examples/").String(),
											Inner: h.String("Examples"),
										},
									},
									&h.A{
										ID:    "rell-run-code",
										Class: "fb-blue",
										Inner: &h.Span{Inner: h.String("Run Code")},
									},
									&h.Label{
										Class: "fb-gray save-code",
										Inner: &h.Input{
											Type:  "submit",
											Value: "Save Code",
										},
									},
									&h.Select{
										ID:   "rell-view-mode",
										Name: "view-mode",
										Inner: &h.Frag{
											&h.Option{
												Inner:    h.String("Website"),
												Selected: c.ViewMode == context.Website,
												Value:    string(context.Website),
												Data: map[string]interface{}{
													"url": c.URL(example.URL).String(),
												},
											},
											&h.Option{
												Inner:    h.String("Canvas"),
												Selected: c.ViewMode == context.Canvas,
												Value:    string(context.Canvas),
												Data: map[string]interface{}{
													"url": c.CanvasURL(example.URL),
												},
											},
											&h.Option{
												Inner:    h.String("Page Tab"),
												Selected: c.ViewMode == context.PageTab,
												Value:    string(context.PageTab),
												Data: map[string]interface{}{
													"url": c.PageTabURL(example.URL),
												},
											},
										},
									},
								},
							},
						},
					},
					&h.Div{ID: "jsroot"},
				},
			},
			&h.Div{
				ID: "log-container",
				Inner: &h.Frag{
					&h.Div{
						Class: "controls",
						Inner: &h.Button{
							ID:    "rell-log-clear",
							Inner: h.String("Clear"),
						},
					},
					&h.Div{ID: "log"},
				},
			},
		},
	}
}

func renderList(context *context.Context, db *examples.DB) *view.Page {
	categories := &h.Frag{}
	for _, category := range db.Category {
		if !category.Hidden {
			categories.Append(renderCategory(context, category))
		}
	}
	return &view.Page{
		Title: "Examples",
		Class: "examples",
		Body: &h.Frag{
			&h.H1{Inner: h.String("Examples")},
			categories,
		},
	}
}

func renderCategory(c *context.Context, category *examples.Category) h.HTML {
	li := &h.Frag{}
	for _, example := range category.Example {
		li.Append(&h.Li{
			Inner: &h.A{
				HREF:  c.URL(example.URL).String(),
				Inner: h.String(example.Name),
			},
		})
	}
	return &h.Frag{
		&h.H2{Inner: h.String(category.Name)},
		&h.Ul{Inner: li},
	}
}

func renderEnvSelector(c *context.Context, example *examples.Example) h.HTML {
	if !c.IsEmployee {
		return nil
	}
	frag := &h.Frag{}
	for title, value := range envOptions {
		ctxCopy := c.Copy()
		ctxCopy.Env = value
		frag.Append(&h.Option{
			Inner:    h.String(title),
			Selected: c.Env == value,
			Value:    value,
			Data: map[string]interface{}{
				"url": ctxCopy.ViewURL(example.URL),
			},
		})
	}
	return &h.Select{
		ID:    "rell-env",
		Name:  "env",
		Inner: frag,
	}
}

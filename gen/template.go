package gen

import (
	"io"
	"text/template"

	"github.com/golang/glog"
	"google.golang.org/protobuf/compiler/protogen"
)

// Options are the options to set for rendering the template.
type Options struct {
	Partial            bool
	Multiline          bool
	EnumsAsInts        bool
	EmitDefaults       bool
	OrigName           bool
	AllowUnknownFields bool
}

// This function is called with a param which contains the entire definition of a method.
func ApplyTemplate(w io.Writer, f *protogen.File, opts Options) error {

	if err := headerTemplate.Execute(w, tplHeader{
		File: f,
	}); err != nil {
		return err
	}

	return applyMessages(w, f.Messages, opts)
}

func applyMessages(w io.Writer, msgs []*protogen.Message, opts Options) error {
	for _, m := range msgs {

		if m.Desc.IsMapEntry() {
			glog.V(2).Infof("Skipping %s, mapentry message", m.GoIdent.GoName)
			continue
		}

		glog.V(2).Infof("Processing %s", m.GoIdent.GoName)
		if err := messageTemplate.Execute(w, tplMessage{
			Message: m,
			Options: opts,
		}); err != nil {
			return err
		}

		if err := applyMessages(w, m.Messages, opts); err != nil {
			return err
		}

	}

	return nil
}

type tplHeader struct {
	*protogen.File
}

type tplMessage struct {
	*protogen.Message
	Options
}

var (
	headerTemplate = template.Must(template.New("header").Parse(`
// Code generated by protoc-gen-go-json. DO NOT EDIT.
// source: {{.Proto.Name}}

package {{.GoPackageName}}

import (
	"google.golang.org/protobuf/encoding/protojson"
)
`))

	messageTemplate = template.Must(template.New("message").Parse(`
// MarshalJSON implements json.Marshaler
func (msg *{{.GoIdent.GoName}}) MarshalJSON() ([]byte,error) {
	return protojson.MarshalOptions {
		UseEnumNumbers: {{.EnumsAsInts}},
		EmitUnpopulated: {{.EmitDefaults}},
		UseProtoNames: {{.OrigName}},
		AllowPartial: {{.Partial}},
		{{- if .Multiline}}
		Multiline: true,
		Indent: "\t",
		{{- end}}
	}.Marshal(msg)
}

// UnmarshalJSON implements json.Unmarshaler
func (msg *{{.GoIdent.GoName}}) UnmarshalJSON(b []byte) error {
	return protojson.UnmarshalOptions {
		DiscardUnknown: {{.AllowUnknownFields}},
	}.Unmarshal(b, msg)
}
`))
)

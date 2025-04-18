package templatex

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

const name = "_"

// T is a template helper.
type T struct {
	t      *template.Template
	buffer *bytes.Buffer
	fm     template.FuncMap
}

// New creates a new template helper.
func New() *T {
	t := template.New(name)
	return &T{
		t:      t,
		buffer: bytes.NewBuffer(nil),
		fm:     funcMap,
	}
}

func (t *T) AppendFuncMap(fm template.FuncMap) {
	for k, v := range fm {
		t.fm[k] = v
	}
}

// MustParse parses the template.
func (t *T) MustParse(text string) *T {
	t.t.Funcs(t.fm)
	_, err := t.t.Parse(text)
	must(err)
	return t
}

// MustExecute executes the template.
func (t *T) MustExecute(data interface{}) *T {
	t.buffer.Reset()
	must(t.t.Execute(t.buffer, data))
	return t
}

// MustSaveAs saves the template to the given filename, it will overwrite the file if it exists.
func (t *T) MustSaveAs(filename string, formatCode bool) {
	var data []byte
	var err error
	if formatCode {
		data, err = source(t.buffer.Bytes())
		if err != nil {
			extension := filepath.Ext(filename)
			errorFilename := strings.TrimSuffix(filename, extension) + ".error" + extension
			ioutil.WriteFile(errorFilename, t.buffer.Bytes(), 0o644)
			must(err)
		}
	} else {
		data = t.buffer.Bytes()
	}
	must(ioutil.WriteFile(filename, data, 0o666))
}

// MustSave saves the template to the given filename, it will do nothing if it exists.
func (t *T) MustSave(filename string, format bool) {
	_, err := os.Stat(filename)
	if err != nil {
		t.MustSaveAs(filename, format)
	}
}

func (t *T) Write(writer io.Writer, formatCode bool) {
	var data []byte
	var err error
	if formatCode {
		data, err = source(t.buffer.Bytes())
		if err != nil {
			fmt.Printf("%+v\n", string(t.buffer.Bytes()))
			must(err)
		}
	} else {
		data = t.buffer.Bytes()
	}
	writer.Write(data)
}

func must(err error) {
	if err == nil {
		return
	}
	fmt.Println(err.Error())
	os.Exit(1)
}

// source formats go code and imports.
func source(data []byte) ([]byte, error) {
	ret, err := format.Source(data)
	if err != nil {
		return nil, err
	}

	return imports.Process("", ret, nil)
}

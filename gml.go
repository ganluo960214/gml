package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/go-playground/validator/v10"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
)

/*
env
*/
var (
	GOFILE    = os.Getenv("GOFILE")
	GOPACKAGE = os.Getenv("GOPACKAGE")
)

/*
validator
*/
var (
	validate = validator.New()
)

/*
flags
*/
var (
	flags = struct {
		Type     string `validate:"required"`
		FileName string
	}{}
)

func init() {
	flag.StringVar(&flags.Type, "type", "", "type name; must be set")
	flag.StringVar(&flags.FileName, "file-name", "", "newly generated file name; default as \"-type_gml.go\"") // flag -file-name-type
}

/*
flags usage
*/
func usage() {
	log.Println("usage: go:generate gml -type=example -file-name=u_can_set_file_name_or_by_default__-type_gml.go")
	log.Println("")
	log.Println("-type data type")
	log.Println("\t data type,must be set")
	log.Println("-file-name newly generated file name, default as \"-type_gml.go\" ")
	log.Println("\t file name of the generated file")
}

/*
init log and flag
*/
func init() {
	log.SetFlags(0)
	log.SetPrefix("gml: ")
	flag.Usage = usage
	flag.Parse()
}

/*
file template
*/
type FileTemplateContent struct {
	TYPE           string
	PACKAGE        string
	MAPPER_CONTENT map[string]string
	LIST_CONTENT   []string
}

const FileTemplate = `// Code generated by "gml -type={{.TYPE}}"; DO NOT EDIT.

package {{.PACKAGE}}

var (
	{{.TYPE}}Mapper = map[{{.TYPE}}]string{ 
		{{range $k,$v := .MAPPER_CONTENT}}{{$k}}:"{{$v}}",{{end}}
	}
	{{.TYPE}}List = []{{.TYPE}}{
		{{stringsJoin .LIST_CONTENT ","}},
	}
)
`

func main() {
	// flags validator
	if err := validate.Struct(&flags); err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(GOFILE); os.IsNotExist(err) {
		log.Fatal(err)
	}

	// ast analysis
	fset := token.NewFileSet()
	astF, err := parser.ParseFile(fset, GOFILE, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	cmap := ast.NewCommentMap(fset, astF, astF.Comments)

	// file template content
	ftc := FileTemplateContent{
		TYPE:           flags.Type,
		PACKAGE:        GOPACKAGE,
		MAPPER_CONTENT: map[string]string{},
		LIST_CONTENT:   []string{},
	}

	// check have const/var variable
	valueSpecsZero := true

	for k := range cmap {
		if vSpec, ok := k.(*ast.ValueSpec); ok {
			valueSpecsZero = false
			// convert to *ast.Ident to get name
			ident, ok := vSpec.Type.(*ast.Ident)
			if !ok {
				log.Fatal("I don't know why failed, please let me know, report a issue with your need generate data!!!") // todo I don't know why failed.
			}
			if ident.Name != flags.Type {
				continue
			}

			name := vSpec.Names[0].Name
			comment := strings.TrimSpace(vSpec.Comment.Text())
			ftc.LIST_CONTENT = append(ftc.LIST_CONTENT, name)
			ftc.MAPPER_CONTENT[name] = comment

		}
	}
	if valueSpecsZero {
		log.Fatal("no values")
	}

	// file content container
	b := bytes.NewBuffer([]byte{})

	// new file template
	t, err := template.New("").Funcs(template.FuncMap{
		"stringsJoin": strings.Join,
	}).Parse(FileTemplate)
	if err != nil {
		log.Fatal(err)
	}
	// generate file content write to file content container
	if err := t.Execute(b, ftc); err != nil {
		log.Fatal(err)
	}

	// file name
	fileName := fmt.Sprintf("%s_gml.go", flags.Type)
	if flags.FileName != "" {
		fileName = flags.FileName
	}

	content, err := format.Source(b.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(fileName, content, 0644); err != nil {
		log.Fatal(err)
	}

}

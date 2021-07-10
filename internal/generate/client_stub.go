package generate

import (
	_ "embed"
	"os"
	"path/filepath"
	"text/template"

	"github.com/isbang/lazy/internal/parser"
)

//go:embed tmpl/client.tmpl
var clientStubCode string

var clientStubTmpl *template.Template

func init() {
	clientStubTmpl = template.Must(template.New("").Parse(clientStubCode))
}

func ClientStub(jobpath string, outdir string) error {
	modname, err := parser.LoadModuleName()
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(outdir, "client.go"))
	if err != nil {
		return err
	}
	defer out.Close()

	structs, err := parser.LoadStructName(jobpath)
	if err != nil {
		return err
	}

	if err := clientStubTmpl.Execute(out, clientArgs{
		RootPackageName: modname,
		JobNames:        structs,
	}); err != nil {
		return err
	}

	return nil
}

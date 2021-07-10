package generate

import (
	_ "embed"
	"os"
	"path/filepath"
	"text/template"

	"github.com/isbang/lazy/internal/parser"
)

//go:embed tmpl/server.tmpl
var serverStubCode string

var serverStubTmpl *template.Template

func init() {
	serverStubTmpl = template.Must(template.New("").Parse(serverStubCode))
}

func ServerStub(jobpath string, outdir string) error {
	modname, err := parser.LoadModuleName()
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(outdir, "server.go"))
	if err != nil {
		return err
	}
	defer out.Close()

	structs, err := parser.LoadStructName(jobpath)
	if err != nil {
		return err
	}

	if err := serverStubTmpl.Execute(out, serverArgs{
		RootPackageName: modname,
		JobNames:        structs,
	}); err != nil {
		return err
	}

	return nil
}

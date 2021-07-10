package generate

import (
	_ "embed"
	"os"
	"text/template"
)

//go:embed tmpl/job_stub.tmpl
var jobStub string

var (
	jobStubTmpl *template.Template
)

func init() {
	jobStubTmpl = template.Must(template.New("").Parse(jobStub))
}

func JobStub(path, job string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := jobStubTmpl.Execute(f, jobstubStructArgs{
		Name: job,
	}); err != nil {
		return err
	}

	return nil
}

package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v3"
)

type moduleOpt struct {
	Type        *string
	Description *any
	Choices     *[]string
	Elements *string
	Required *bool
	Aliases  *[]string // todo use aliases?
}

func (x moduleOpt) IntoDescription() (t []string) {
	t = []string{}
	if x.Description == nil {
		return
	}

	ty := reflect.TypeOf(*x.Description)
	switch ty.String() {
	case "[]interface {}":
		y := *x.Description
		for _, v := range y.([]interface{}) {
			t = append(t, v.(string))
		}
	case "string":
		y := *x.Description
		t = append(t, y.(string))
	default:
		log.Printf("%+v", ty)
	}

	return
}

func (x moduleOpt) IntoType() (t string) {
	var outer func(s *string) string
	outer = func(s *string) (t string) {
		switch {
		case s == nil:
			t = "String"

		case *s == "bool": // todo default yes???
			t = "Boolean"
		case *s == "int":
			t = "Int"
		case *s == "list":
			inner := outer(x.Elements)
			t = "Listing<" + inner + ">"

		case *s == "str":
			switch {
			case x.Choices != nil:
				t = t + "("
				for i, c := range *x.Choices {
					if i > 0 {
						t = t + "|"
					}
					t = t + `"` + c + `"`
				}
				t = t + ")"
			default:
				t = "String"
			}

		case *s == "path": // todo path
			fallthrough
		case *s == "raw": // todo raw
			fallthrough
		default:
			t = "String"
		}
		return t
	}
	t = outer(x.Type)

	//

	if x.Required == nil || !*x.Required {
		t = t + "?"
	}

	return
}

type ansibleModule struct {
	Module           string
	ShortDescription string `yaml:"short_description"`
	Options          map[string]moduleOpt
}

func (x ansibleModule) ToCamel() string {
    return strcase.ToCamel(x.Module)
}

func main() {
	if err := mainErr(); err != nil {
		log.Fatal(err)
	}
}

func mainErr() error {
	if err := readModules(
		"ansible.builtin",
		"./ansible/lib/ansible/modules",
		"./src",
	); err != nil {
		return err
	}

	if err := readModules(
		"community.general",
		"./community.general/plugins/modules",
		"./src",
	); err != nil {
		return err
	}

	return nil
}

func readModules(
	ansibleModName string,
	srcDir string,
	outDir string,
) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	var modules []ansibleModule

	for _, entry := range entries {
		// deprecated module
		if entry.Name() == "webfaction_site.py" {
			continue
		}

		if entry.Name() == "__init__.py" {
			continue
		}
		if entry.IsDir() {
			continue
		}
		if !entry.Type().IsRegular() {
			continue
		}

		//

		srcPath := srcDir + "/" + entry.Name()
		log.Printf("parse ast %s", srcPath)
		cmd := exec.CommandContext(context.TODO(), "./read_doc.py", srcPath)
		cmd.Stderr = os.Stderr
		outBytes, err := cmd.Output()
		if err != nil {
			return err
		}

        if strings.HasPrefix(string(outBytes), "TODO_NODOC") {
            log.Printf("TODO_NODOC %s", srcPath)
            continue
        }
        if strings.HasPrefix(string(outBytes), "TODO_EXCEPTION") {
            log.Printf("TODO_EXCEPTION %s", srcPath)
            continue
        }

		var m ansibleModule
		if err := yaml.Unmarshal(outBytes, &m); err != nil {
			return err
		}

		modules = append(modules, m)
	}

    pklModName := strcase.ToCamel(ansibleModName)

	file, err := os.Create(outDir + "/" + pklModName + ".pkl")
	if err != nil {
		return err
	}
	defer file.Close()

	t, err := template.New(pklModName).Parse(`module ` + pklModName + `

import "./Playbook.pkl"

`)
	if err := t.Execute(file, nil); err != nil {
		return err
	}

	// todo "free-form" option
	// todo IntoDescription crashes
	for _, m := range modules {
		t, err := template.New(m.Module).Funcs(template.FuncMap{
			"IntoProperty": func(s string) (z string) {
				z = s
				switch s {
				case "hidden":
					z = "`hidden`"
				case "local":
					z = "`local`"
				case "switch":
					z = "`switch`"
				case "record":
					z = "`record`"
				case "external":
					z = "`external`"
				case "override":
					z = "`override`"
				case "delete":
					z = "`delete`"
				}
				return
			},
		}).Parse(`// ` + m.ShortDescription + `
class ` + m.ToCamel() + `Options {
    {{ range $key, $value := .module.Options }}
    {{- range $i, $v :=  $value.IntoDescription }}
    {{- if lt $i 0 }}
    // {{ . }}
    {{- end }}
    {{- end }}
    {{- if eq $key "free-form" }}
    // {{ IntoProperty $key }}: {{ $value.IntoType }}
    {{ else }}
    {{ IntoProperty $key }}: {{ $value.IntoType }}
    {{- end }}
    {{ end }}
}

// Task class for ` + m.Module + `
class ` + m.ToCamel() + ` extends Playbook.Task {

    hidden options: ` + m.ToCamel() + `Options

    ` + "`" + ansibleModName + "." + m.Module + "`" + ": " + m.ToCamel() + `Options?

    function into(): ` + m.ToCamel() + ` = this
        .toMap()
        .put("` + ansibleModName + "." + m.Module + `", this.options)
        .toTyped(` + m.ToCamel() + `)

}

`)
		if err != nil {
			return err
		}

		if err := t.Execute(file, map[string]interface{}{
			"module": m,
		}); err != nil {
			return err
		}
	}

	return nil
}

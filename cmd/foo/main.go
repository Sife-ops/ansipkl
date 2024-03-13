package main

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type moduleOpt struct {
	Type     *string
	Elements *string
	Required *bool
	Aliases  *[]string
}

func (x moduleOpt) IntoType() (t string) {
	var outer func(s *string) string
	outer = func(s *string) (t string) {
		switch {
		case s == nil:
			t = "String"

		case *s == "bool":
			t = "Boolean"
		case *s == "int":
			t = "Int"
		case *s == "list":
			inner := outer(x.Elements)
			t = "Listing<" + inner + ">"

		case *s == "str":
            fallthrough
		case *s == "path":
            fallthrough
		default:
			t = "String"
		}
		return t
	}
	t = outer(x.Type)

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

func main() {
	if err := mainErr(); err != nil {
		log.Fatal(err)
	}
}

func mainErr() error {
	basename := "./ansible/lib/ansible/modules"

	entries, err := os.ReadDir(basename)
	if err != nil {
		return err
	}

	var modules []ansibleModule

	for i, entry := range entries {
		if entry.Name() == "__init__.py" {
			continue
		}
		if entry.IsDir() {
			continue
		}
		if !entry.Type().IsRegular() {
			continue
		}

		// if entry.Name() != "add_host.py" {
		// 	continue
		// }
		if i > 2 {
			continue
		}
		log.Print(entry.Name())

		file, err := os.Open(basename + "/" + entry.Name())
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		docStart := false
		docEnd := false
		buf := new(bytes.Buffer)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "DOCUMENTATION") {
				docStart = true
				continue
			}
			if scanner.Text() == "'''" && docStart {
				docEnd = true
			}
			if !docStart || docEnd {
				continue
			}
			if _, err := buf.Write([]byte(scanner.Text() + "\n")); err != nil {
				return err
			}
		}

		var m ansibleModule
		if err := yaml.Unmarshal(buf.Bytes(), &m); err != nil {
			return err
		}

		modules = append(modules, m)
	}

	file, err := os.Create("./src/builtin.pkl")
	if err != nil {
		return err
	}
	defer file.Close()

	t, err := template.New("builtin").Parse(`module builtin

import "./todoname0.pkl"
`)
	if err := t.Execute(file, nil); err != nil {
		return err
	}

	for _, m := range modules {
		t, err := t.Parse(`
//

class ` + m.Module + `_options {
    {{ range $key, $value := .mod.Options }}
    {{ $key }}: {{ $value.IntoType }}
    {{ end }}
}

class ` + m.Module + ` extends todoname0.task {
    hidden options: ` + m.Module + `_options

    ` + "`" + `ansible.builtin.` + m.Module + "`" + `: ` + m.Module + `_options?

    function into(): ` + m.Module + ` = this
        .toMap()
        .put("ansible.builtin.` + m.Module + `", this.options)
        .toTyped(` + m.Module + `)
}
`)
		if err != nil {
			return err
		}

		if err := t.Execute(file, map[string]interface{}{
			"mod": m,
		}); err != nil {
			return err
		}
	}

	return nil
}

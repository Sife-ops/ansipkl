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

// todo descriptions, docs
type moduleOpt struct {
	Type     *string
	Choices  *[]string
	Default  *string
	Elements *string
	Required *bool
	Aliases  *[]string // todo use aliases?
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

	for _, entry := range entries {
		if entry.Name() == "__init__.py" {
			continue
		}

        // very bad
		log.Print(entry.Name())

		if entry.Name() == "deb822_repository.py" {
			continue
		}
		if entry.Name() == "dnf.py" {
			continue
		}
		if entry.Name() == "dnf5.py" {
			continue
		}
		if entry.Name() == "find.py" {
			continue
		}
		if entry.Name() == "get_url.py" {
			continue
		}
		if entry.Name() == "git.py" {
			continue
		}
		if entry.Name() == "include_vars.py" {
			continue
		}
		if entry.Name() == "iptables.py" {
			continue
		}
		if entry.Name() == "package_facts.py" {
			continue
		}
		if entry.Name() == "reboot.py" {
			continue
		}
		if entry.Name() == "setup.py" {
			continue
		}
		if entry.Name() == "unarchive.py" {
			continue
		}
		if entry.Name() == "uri.py" {
			continue
		}
		if entry.Name() == "wait_for.py" {
			continue
		}
		if entry.Name() == "yum.py" {
			continue
		}
        //

		if entry.IsDir() {
			continue
		}
		if !entry.Type().IsRegular() {
			continue
		}

		file, err := os.Open(basename + "/" + entry.Name())
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		docStart := false
		docEnd := false
		buf := new(bytes.Buffer)
        // iii := 0
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
            // xxx++
            // log.Printf("%d | %s", iii, scanner.Text())
			// if _, err := buf.Write([]byte(scanner.Text() + "\n")); err != nil {
			if _, err := buf.Write(scanner.Bytes()); err != nil {
				return err
			}
			if _, err := buf.Write([]byte("\n")); err != nil {
				return err
			}
		}

		var m ansibleModule
		if err := yaml.Unmarshal(buf.Bytes(), &m); err != nil {
			return err
		}

		if docStart {
			modules = append(modules, m)
		}
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

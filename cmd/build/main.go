package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v3"
    "github.com/Sife-ops/ansipkl/lib"
)

func main() {
	if err := mainErr(); err != nil {
		log.Fatal(err)
	}
}

func mainErr() error {
    //
	t, err := lib.TplPlaybook()
	if err != nil {
		return err
	}

	file, err := os.Create("./src/Playbook.pkl")
	if err != nil {
		return err
	}
	defer file.Close()

	classBase, err := readClass("./ansible/lib/ansible/playbook/base.py", "Base")
	if err != nil {
		return err
	}
	classPlay, err := readClass("./ansible/lib/ansible/playbook/play.py", "Play")
	if err != nil {
		return err
	}
	classTask, err := readClass("./ansible/lib/ansible/playbook/task.py", "Task")
	if err != nil {
		return err
	}
	classBlock, err := readClass("./ansible/lib/ansible/playbook/block.py", "Block")
	if err != nil {
		return err
	}

	if err := t.Execute(file, map[string]interface{}{
		"classBase":  classBase,
		"classPlay":  classPlay,
		"classTask":  classTask,
		"classBlock": classBlock,
	}); err != nil {
		return err
	}

	//
	if err := readModules("ansible.builtin", "./ansible/lib/ansible/modules"); err != nil {
		return err
	}
	if err := readModules("community.general", "./community.general/plugins/modules"); err != nil {
		return err
	}
	if err := readModules("community.docker", "./community.docker/plugins/modules"); err != nil {
		return err
	}

	return nil
}

func readClass(s0 string, s1 string) (sa []string, err error) {
    sa = []string{}
    cmd := exec.CommandContext(context.TODO(), "./read_class.py", s0, s1)
    cmd.Stderr = os.Stderr
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return
    }
    if err = cmd.Start(); err != nil {
        return
    }

    scanner := bufio.NewScanner(stdout)
    for scanner.Scan() {
        // overrides
        text := scanner.Text()
        switch s1 {
        case "Play":
            switch {
            case strings.HasPrefix(text, "hosts:"):
                continue
            case strings.HasPrefix(text, "tasks:"):
                continue
            case strings.HasPrefix(text, "post_tasks:"):
                continue
            case strings.HasPrefix(text, "pre_tasks:"):
                continue
            case strings.HasPrefix(text, "gather_subset:"):
                continue
            case strings.HasPrefix(text, "vars_files:"):
                continue
            case strings.HasPrefix(text, "vars_prompt:"):
                continue
            case strings.HasPrefix(text, "roles:"):
                continue
            case strings.HasPrefix(text, "handlers:"):
                continue
            case strings.HasPrefix(text, "serial:"):
                continue
            }
        case "Block":
            switch {
            case strings.HasPrefix(text, "block:"):
                continue
            case strings.HasPrefix(text, "rescue:"):
                continue
            case strings.HasPrefix(text, "always:"):
                continue
            }
        case "Task":
            switch {
            case strings.HasPrefix(text, "changed_when:"):
                continue
            case strings.HasPrefix(text, "failed_when:"):
                continue
            case strings.HasPrefix(text, "loop:"):
                continue
            case strings.HasPrefix(text, "until:"):
                continue
            }
        }
        sa = append(sa, text)
    }

    if err = cmd.Wait(); err != nil {
        return
    }

    return
}

func readModules(ansibleModName string, srcDir string) error {
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

	file, err := os.Create("./src/" + pklModName + ".pkl")
	if err != nil {
		return err
	}
	defer file.Close()

	t, err := template.New(pklModName).Parse(
`module ` + pklModName + `

import "./Playbook.pkl"

`)
	if err := t.Execute(file, nil); err != nil {
		return err
	}

	for i, m := range modules {
        t, err := lib.TplModule(ansibleModName)
        if err != nil {
            return err
        }

		if err := t.Execute(file, map[string]interface{}{
			"moduleIndex": i,
			"module":      m,
		}); err != nil {
			return err
		}
	}

	return nil
}

type ansibleModule struct {
	Module           string
	Description      *any
	ShortDescription string `yaml:"short_description"`
	Options          map[string]moduleOpt
}

func (x ansibleModule) IntoDescription() (t []string) {
	t = []string{}
	if x.Description == nil {
		return
	}

	ty := reflect.TypeOf(*x.Description)
	switch ty.String() {
	case "[]interface {}":
		y := *x.Description
		for _, v := range y.([]interface{}) {
			t = append(t, strings.ReplaceAll(v.(string), "\n", ""))
		}
	case "string":
		y := *x.Description
		t = append(t, strings.ReplaceAll(y.(string), "\n", ""))
	default:
		log.Printf("%+v", ty)
	}

	return
}

func (x ansibleModule) ToCamel() string {
	return strcase.ToCamel(x.Module)
}

type moduleOpt struct {
	Type        *string
	Description *any
	Choices     *[]string
	Elements    *string
	Required    *bool
	Aliases     *[]string // todo use aliases?
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
			t = append(t, strings.ReplaceAll(v.(string), "\n", ""))
		}
	case "string":
		y := *x.Description
		t = append(t, strings.ReplaceAll(y.(string), "\n", ""))
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
			t = "Any"

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
			t = "String"

		default:
			t = "Any"
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

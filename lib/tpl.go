package lib

import "text/template"

// todo handle "free-form" option
func TplPlaybook() (*template.Template, error) {
	t, err := template.New("Playbook").Parse(
		`module Playbook

abstract class Base {
    {{- range .classBase }}
    {{ . }}
    {{- end }}
}

abstract class Plays extends Base {}

function NewPlaybook(x: Listing<Plays>): Listing<Plays> = x

class Play extends Plays {
    hosts: String?
    tasks: Listing<Tasks>?
    post_tasks: Listing<Tasks>?
    pre_tasks: Listing<Tasks>?
    gather_subset: Listing<String>?
    vars_files: Listing<String>?
    vars_prompt: Listing<Dynamic>?
    roles: Listing<Role>?
    handlers: Listing<Tasks>?
    serial: (String|Int|Listing<String|Int>)?

    function ModuleDefaults(x: Listing<Tasks>): Play = this
        |> new Mixin {
            module_defaults = x
                .toList()
                .toMap((y) -> y.GetModuleName(), (y) -> y.GetModuleOptions())
                .toDynamic()
        }

    {{ range .classPlay }}
    {{ . }}
    {{- end }}

    // base classes
    {{- block "Taggable" . }} {{ end }}
    {{- block "CollectionSearch" . }} {{ end }}
}

class PlaybookInclude extends Plays {
    ` + "`ansible.builtin.import_playbook`" + `: String?
    vars: Dynamic?

    {{ block "Conditional" . }} {{ end }}
    {{- block "Taggable" . }} {{ end }}
}

class ImportPlaybook extends Base {
    import_playbook: String?
    vars_val: Dynamic?

    function Include(): PlaybookInclude = this
        .toMap()
        .put("ansible.builtin.import_playbook", this.import_playbook)
        .put("vars", this.vars_val)
        .toTyped(PlaybookInclude)

    {{ block "Conditional" . }} {{ end }}
    {{- block "Taggable" . }} {{ end }}
}

abstract class Tasks extends Base {}

function NewRole(x: Listing<Tasks>): Listing<Tasks> = x

abstract class Task extends Tasks {
    changed_when: Any?
    failed_when: Any?
    loop: (String|Listing<String>)?
    until: (String|Listing<String>)?

    {{ range .classTask }}
    {{ . }}
    {{- end }}

    {{ block "Conditional" . }} {{ end }}
    {{- block "Taggable" . }} {{ end }}
    {{- block "CollectionSearch" . }} {{ end }}
    {{- block "Notifiable" . }} {{ end }}
    {{- block "Delegatable" . }} {{ end }}
}

abstract class TaskBuilder extends Base {
    options_mixin: Mixin?

    changed_when: Any?
    failed_when: Any?
    loop: (String|Listing<String>)?
    until: (String|Listing<String>)?

    {{ range .classTask }}
    {{ . }}
    {{- end }}

    {{ block "Conditional" . }} {{ end }}
    {{- block "Taggable" . }} {{ end }}
    {{- block "CollectionSearch" . }} {{ end }}
    {{- block "Notifiable" . }} {{ end }}
    {{- block "Delegatable" . }} {{ end }}
}

class Block extends Tasks {
    block: Listing<Tasks>?
    rescue: Listing<Tasks>?
    always: Listing<Tasks>?

    {{- range .classBlock }}
    {{ . }}
    {{- end }}

    {{ block "Conditional" . }} {{ end }}
    {{- block "CollectionSearch" . }} {{ end }}
    {{- block "Taggable" . }} {{ end }}
    {{- block "Notifiable" . }} {{ end }}
    {{- block "Delegatable" . }} {{ end }}
}

class Handler extends Task {
    listen: Dynamic?
}

class Role extends Base {
    role: String

    {{ block "Conditional" . }} {{ end }}
    {{- block "CollectionSearch" . }} {{ end }}
    {{- block "Taggable" . }} {{ end }}
}

`)
	if err != nil {
		return nil, err
	}

	return t.Parse(
		`{{ define "Taggable" }}
    tags: (String|Listing<String>)?
    {{- end }}

    {{ define "CollectionSearch" }}
    collections: (String|Listing<String>)?
    {{- end }}

    {{ define "Conditional" }}
    ` + "`" + "when" + "`" + `: String?
    {{- end }}

    {{ define "Notifiable" }}
    notify: (String|Listing<String>)?
    {{- end }}

    {{ define "Delegatable" }}
    delegate_to: String?
    delegate_facts: Boolean?
    {{- end }}`)
}

func TplModule(s string) (*template.Template, error) {
	funcMap := template.FuncMap{
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
	}

	return template.
		New(s).
		Funcs(funcMap).
		Parse(
			`/// {{ .module.ShortDescription }}
{{- range $.module.IntoDescription }}
/// {{ . }}
{{- end }}
class {{ .module.ToCamel }}Options {
    {{- range $key, $value := .module.Options }}
    {{- if eq $key "free-form" }}

    // {{ IntoProperty $key }}: {{ $value.IntoType }}

    {{ else }}

    {{- range $value.IntoDescription }}
    /// {{ . }}
    {{- end }}
    {{ IntoProperty $key }}: {{ $value.IntoType }}

    {{- end }}
    {{- end }}
}

/// Task class for {{ .module.Module }}
class {{ .module.ToCamel }}Task extends Playbook.Task {

    ` + "`" + s + ".{{ .module.Module }}" + "`" + `: Dynamic

    function GetModuleName(): String = "` + s + `.{{ .module.Module }}"
    function GetModuleOptions(): Dynamic = this.` + "`" + s + ".{{ .module.Module }}" + "`" + `
}

/// TaskBuilder class for {{ .module.Module }}
class {{ .module.ToCamel }} extends Playbook.TaskBuilder {
    /// Options for ` + s + `.{{ .module.Module }}
    options: {{ .module.ToCamel }}Options?

    function Task(): {{ .module.ToCamel }}Task = this
        .toMap()
        .put("` + s + `.{{ .module.Module }}", (this.options.ifNonNull((it) -> it.toDynamic()) ?? new Dynamic {})
            |> (this.options_mixin ?? new Mixin {}))
        .toTyped({{ .module.ToCamel }}Task)
}

`)
}


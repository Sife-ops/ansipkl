# ansipkl

![example](https://i.imgur.com/EGdHvsa.png)

ansipkl provides `pkl` types for Ansible plays, tasks, and modules, etc.

## add to project

```pkl
amends "pkl:Project"

dependencies {
    ["ansipkl"] { 
        uri = "package://github.com/Sife-ops/ansipkl/releases/download/v0.0.1/ansipkl@0.0.1"
    }
}
```

```bash
pkl project resolve
```

## install cli

you must have the `pkl` executable installed

```bash
go install github.com/Sife-ops/ansipkl@latest
```

example `ansipkl.toml`

```toml
[options]
exclude = [
    "foo",
    "^bar/.*"
]
```

## usage

Use `ansipkl` to convert all `.pkl` files to `.yml` recursively.
Check out `/example`.

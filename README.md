# ansipkl

![example](https://i.imgur.com/ATcJ90A.png)

`ansipkl` provides `pkl` types for Ansible plays, tasks, and modules, etc.

## add to project

```pkl
amends "pkl:Project"

dependencies {
    ["ansipkl"] { 
        uri = "package://github.com/Sife-ops/ansipkl/releases/download/v<version>/ansipkl@<version>"
    }
}
```

```bash
pkl project resolve
```

## install cli

The `ansipkl` command depends on the `pkl` executable. You must have `pkl`
installed.

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

Use `ansipkl` to convert all `.pkl` files to `.yml` recursively. Check out
`/example`.

# ansipkl

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

check `/example`

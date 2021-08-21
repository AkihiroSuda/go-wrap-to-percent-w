# go-wrap-to-percent-w: convert `Wrap(err, "foo")` to `Errors("foo: %w", err)`

`go-wrap-to-percent-w` converts legacy [`github.com/pkg/errors.Wrap(err, "foo")`](https://pkg.go.dev/github.com/pkg/errors#Wrap)
to modern Go-native [`fmt.Errors("foo: %w", err)`](https://pkg.go.dev/fmt#Errorf)
introduced in [Go 1.13](https://go.dev/blog/go1.13-errors).

## Conversion rule

| Input                                  | Output                                   |
| -------------------------------------- | ---------------------------------------- |
| `errors.Wrap(err, "foo")`              | `fmt.Errorf("foo: %w", err)`             |
| `errors.Wrapf(err, "foo %s %d", s, d)` | `fmt.Errorf("foo %s %d: %w", s, d, err)` |
| `errors.Errorf("foo %s %d", s,d)`      | `fmt.Errors("foo %s %d", s, d)`          |

TODO: convert `errors.New("foo")` to `stderrors.New("foo")` ?

## Install

```console
go get github.com/AkihiroSuda/go-wrap-to-percent-w
```

## Usage
:warning: Backup your data before conversion

```console
go-wrap-to-percent-w -w *.go
```

Flags:
- `-w`: write result to (source) file instead of stdout (Default: false)
- `-gofmt`: run `gofmt` after conversion (Default: true)

TODO: support specifying package names (`./...`)

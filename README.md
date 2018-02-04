## Commandeer
[![Go Report Card](https://goreportcard.com/badge/github.com/jaffee/commandeer)](https://goreportcard.com/report/github.com/jaffee/commandeer)
[![GoDoc](https://godoc.org/github.com/jaffee/commandeer?status.svg)](https://godoc.org/github.com/jaffee/commandeer)
[![Coverage](http://gocover.io/_badge/github.com/jaffee/commandeer)](https://gocover.io/github.com/jaffee/commandeer)

![Image](https://i.imgur.com/y6GmOGE.png)

Commandeer sets up command line flags based on struct fields and tags.

Do you...
 * like to develop Go apps as libraries with tiny main packages?
 * get frustrated keeping your flags up to date as your code evolves?
 * feel irked by the overlap between comments on struct fields and help strings for flags?
 * hate switching between your app's main and library packages?

You might like Commandeer. See the [godoc](https://godoc.org/github.com/jaffee/commandeer) for detailed usage, or just...

## Try It!
Here's how it works, define your app like so:
```go
package myapp

import "fmt"

type Main struct {
	Num     int    `help:"How many does it take?"`
	Vehicle string `help:"What did they get?"`
}

func NewMain() *Main { return &Main{Num: 5, Vehicle: "jeep"} }

func (m *Main) Run() error {
	if m.Num < 2 || m.Vehicle == "" {
		return fmt.Errorf("Need more gophers and/or vehicles.")
	}
	fmt.Printf("%d gophers stole my %s!\n", m.Num, m.Vehicle)
	return nil
}
```

and your main package:
```go
package main

import (
	"fmt"

	"github.com/jaffee/commandeer"
	"github.com/jaffee/commandeer/examples/myapp"
)

func main() {
	err := commandeer.Run(myapp.NewMain())
	if err != nil {
		fmt.Println(err)
	}
}
```

Now...
```bash
$ ./myapp -h
Usage of ./myapp:
  -num int
    	How many does it take? (default 5)
  -vehicle string
    	What did they get? (default "jeep")

$ ./myapp
5 gophers stole my jeep!
$ ./myapp -num 3 -vehicle horse
3 gophers stole my horse!
```

Notice that Commandeer set up the default values for each flag based on the
values in the struct passed to `Run`.

Commandeer is set up for minimal dependency pollution - it uses only stdlib
dependencies and is a few hundred lines of code itself. You need only import it
from a tiny `main` package (as in the example), and shouldn't need to reference
it anywhere else.

If you aren't allergic to external dependencies, you can also try
`github.com/jaffee/commandeer/cobrafy` which pulls in the excellent [Cobra](https://github.com/spf13/cobra) and
[pflag](https://github.com/spf13/pflag) packages giving you GNU/POSIX style flags and some other nice features
should you care to use them. See the [godoc](https://godoc.org/github.com/jaffee/commandeer/cobrafy), or the [myapp-cobrafy example](https://github.com/jaffee/commandeer/blob/master/examples/myapp/cmd/myapp-cobrafy/main.go).

## Contributing
Yes please!

For small stuff, feel free to submit a PR directly. For larger things,
especially API changes, it's best to make an issue first so it can be discussed.

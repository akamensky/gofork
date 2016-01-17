# GoFork
Awesome forker. 

Forks current Go process with additional arguments and communicate with its stdin, stdout and stderr

### Usage

TODO: Write usage

### Example
TODO: Improve

```
package main

import (
	"github.com/akamensky/gofork"
	"fmt"
	"os"
	"errors"
	"strconv"
)

var fork gofork.Fork

func init() {
	fork = gofork.Fork{}
	if err := fork.Start("--slave"); err != nil {
		panic(err)
	}
}


func main() {
	if fork.IsFork() {
		ChildRun(&fork)
	} else {
		MasterRun(&fork)
	}
}

func MasterRun(fork *gofork.Fork) {
	// Read err and if any, show err and exit
	go func(){
		err := <-fork.Err
		if err != "" {
			fork.Stop()
			panic(errors.New(err))
			os.Exit(1)
		}
	}()
	nums := []int{1,2,3,4,5}
	for _, num := range nums {
		fork.In<- fmt.Sprintf("%d", num)
		fmt.Println(<-fork.Out)
	}
}

func ChildRun(fork *gofork.Fork) {
	for {
		number, err := strconv.Atoi(<-fork.In)
		if err != nil {
			fork.Err<- err.Error()
			os.Exit(1)
		}
		fork.Out<- fmt.Sprintf("%d", number*number)
	}
}
```

# GoFork
Awesome forker. 

Forks current Go process with additional arguments and communicate with its stdin, stdout and stderr

### Usage

Recommended way to use is to activate fork in `init` function:
```
var fork gofork.Fork

func init() {
    fork = gofork.Fork{}
    if err := fork.Start("--child"); err != nil {
        panic(err)
    }
}
```
After fork is spawned it is possible to distinguish whether current code executes in main process or in fork by calling `fork.IsFork()` which returns `true` for forked process and `false` for parent process

Fork structure provides 3 channels `In` for Stdin of forked process, `Out` for Stdout and `Err` for Stderr. Those channels transmit strings as messages between parent and forked process. In parent process you should send messages to `In` and receive from `Out` and `Err`. In forked process you should receive from `In` and send to `Out` and `Err`

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
	if err := fork.Start("--child"); err != nil {
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

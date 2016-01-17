package gofork

import (
	"os/exec"
	"os"
	"reflect"
	"bufio"
	"strings"
	"fmt"
	"io"
)

type Fork struct {
	args []string
	cmd *exec.Cmd
	In chan string
	Out chan string
	Err chan string
	isFork bool
	isAlive bool
	inPipe io.WriteCloser
	outPipe io.ReadCloser
	errPipe io.ReadCloser
}

func (f *Fork) Start(args ...string) error {
	f.initialize(args...)
	err := f.execute()
	return err
}

func (f *Fork) IsFork() bool {
	return f.isFork == true
}

func (f *Fork) IsAlive() bool {
	return f.isAlive == true
}

func (f *Fork) initialize(args ...string) {
	// Save passed arguments (may need them?)
	f.args = append([]string{os.Args[0]}, args...)
	// If current arguments and supplied arguments are equal
	// then this is already fork
	if reflect.DeepEqual(os.Args, f.args) {
		f.cmd = nil
		f.isFork = true
	} else {
		f.cmd = exec.Command(os.Args[0], args...)
		f.isFork = false
	}

	// In either case we should use buffered channels
	f.In = make(chan string, 20)
	f.Out = make(chan string, 20)
	f.Err = make(chan string, 20)
}

func (f *Fork) execute() error {
	if f.isFork {
		// Anything that comes on stdin should go into In channel
		go func() {
			input := bufio.NewReader(os.Stdin)
			for {
				text, err := input.ReadString('\n')
				if err != nil {
					// Parent process exited? So shall we. But with the error since shutdown is not clean
					os.Exit(1)
				}
				f.In <- strings.TrimSpace(text)
			}
		}()
		// Anything that is written into Out channel should go to stdout
		go func() {
			for {
				text := strings.TrimSpace(<-f.Out)
				fmt.Fprintf(os.Stdout, "%s\n", text)
			}
		}()
		// Anything that is written into Err channel should go to stderr
		go func() {
			for {
				text := strings.TrimSpace(<-f.Err)
				fmt.Fprintf(os.Stderr, "%s\n", text)
			}
		}()
	} else {
		// On parent process we first should pipe stdin, stdout and stderr
		f.inPipe, _ = f.cmd.StdinPipe()
		f.outPipe, _ = f.cmd.StdoutPipe()
		f.errPipe, _ = f.cmd.StderrPipe()

		// Next we attempt to start the subprocess
		// if failed return error
		if err := f.cmd.Start(); err != nil {
			return err
		}

		// Set isAlive to true since process is running
		f.isAlive = true

		// Then we start reading/writing threads

		// Anything in In channel goes into subprocess stdin pipe
		go func() {
			for f.isAlive {
				text := strings.TrimSpace(<-f.In)
				if text != "" {
					io.WriteString(f.inPipe, text+"\n")
				}
			}
		}()
		// Anything read from subprocess stdout pipe goes into Out channel
		go func() {
			input := bufio.NewReader(f.outPipe)
			for f.isAlive {
				text, err := input.ReadString('\n')
				if err != nil {
					// Most likely subprocess exited. Set isAlive to false
					f.isAlive = false
					// Avoid detached child
					f.Stop()
					f.inPipe.Close()
					f.outPipe.Close()
					f.errPipe.Close()
					// Now can exit
					return
				}
				f.Out <- strings.TrimSpace(text)
			}
		}()
		// Anything read from subprocess stderr pipe goes into Err channel
		go func() {
			input := bufio.NewReader(f.errPipe)
			for f.isAlive {
				text, err := input.ReadString('\n')
				if err != nil {
					// Most likely subprocess exited. Set isAlive to false
					f.isAlive = false
					// Avoid detached child
					f.Stop()
					f.inPipe.Close()
					f.outPipe.Close()
					f.errPipe.Close()
					// Now can exit
					return
				}
				f.Err <- strings.TrimSpace(text)
			}
		}()
	}

	return nil
}

func (f *Fork) Stop() {
	if f.isFork {
		os.Exit(0)
	} else {
		f.cmd.Process.Kill()
	}
}

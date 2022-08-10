package deadlock

import (
	"fmt"
	"os"
	"runtime"
)

/*
Copyright (c) 2022, Erik Kassubek
All rights reserved.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

/*
Author: Erik Kassubek <erik-kassubek@t-online.de>
Package: deadlock
Project: Bachelor Project at the Albert-Ludwigs-University Freiburg,
	Institute of Computer Science: Dynamic Deadlock Detection in Go
*/

/*
report.go
This file contains functions to report deadlock that were found in any of
the deadlock checks
*/

// colors for deadlock messages
const (
	purple = "\033[1;35m%s\033[0m"
	red    = "\033[1;31m%s\033[0m"
	blue   = "\033[0;36m%s\033[0m"
)

// report if double locking is detected
//  Args:
//   m (mutexInt): mutex on which double locking was detected
//  Returns:
//   nil
func reportDeadlockDoubleLocking(m mutexInt) {
	fmt.Fprintf(os.Stderr, red, "DEADLOCK (DOUBLE LOCKING)\n\n")

	// print information about the involved lock
	fmt.Fprintf(os.Stderr, purple, "Initialization of lock involved in deadlock:\n\n")
	context := *m.getContext()
	fmt.Fprintln(os.Stderr, context[0].file, context[0].line)
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, purple, "Calls of lock involved in deadlock:\n\n")
	for i, call := range context {
		if i == 0 {
			continue
		}
		fmt.Fprintln(os.Stderr, call.file, call.line)
	}
	_, file, line, _ := runtime.Caller(4)
	fmt.Fprintln(os.Stderr, file, line)
	fmt.Fprintf(os.Stderr, "\n\n")
}

// report a found deadlock
//  Args:
//   stack (*depStack) stack which represents the found cycle
//  Returns:
//   nil
func reportDeadlock(stack *depStack) {
	fmt.Fprintf(os.Stderr, red, "POTENTIAL DEADLOCK\n\n")

	// print information about the locks in the circle
	fmt.Fprintf(os.Stderr, purple, "Initialization of locks involved in potential deadlock:\n\n")
	for cl := stack.stack.next; cl != nil; cl = cl.next {
		for _, c := range *cl.depEntry.mu.getContext() {
			if c.create {
				fmt.Fprintln(os.Stderr, c.file, c.line)
			}
		}
	}

	// print information if call stacks were collected
	if opts.collectCallStack {
		fmt.Fprintf(os.Stderr, purple, "\nCallStacks of Locks involved in potential deadlock:\n\n")
		for cl := stack.stack.next; cl != nil; cl = cl.next {
			cont := *cl.depEntry.mu.getContext()
			fmt.Fprintf(os.Stderr, blue, "CallStacks for lock created at: ")
			fmt.Fprintf(os.Stderr, blue, cont[0].file)
			fmt.Fprintf(os.Stderr, blue, ":")
			fmt.Fprintf(os.Stderr, blue, fmt.Sprint(cont[0].line))
			fmt.Fprintf(os.Stderr, "\n\n")
			for i, c := range cont {
				if i != 0 {
					fmt.Fprint(os.Stderr, c.callStacks)
				}
			}
		}
	} else {
		// print information if only caller information were selected
		fmt.Fprintf(os.Stderr, purple, "\nCalls of locks involved in potential deadlock:\n\n")
		for cl := stack.stack.next; cl != nil; cl = cl.next {
			for i, c := range *cl.depEntry.mu.getContext() {
				if i == 0 {
					fmt.Fprintf(os.Stderr, blue, "Calls for lock created at: ")
					fmt.Fprintf(os.Stderr, blue, c.file)
					fmt.Fprintf(os.Stderr, blue, ":")
					fmt.Fprintf(os.Stderr, blue, fmt.Sprint(c.line))
					fmt.Fprintf(os.Stderr, "\n")
				} else {
					fmt.Fprintln(os.Stderr, c.file, c.line)
				}
			}
			fmt.Fprintln(os.Stderr, "")
		}
	}
	fmt.Fprintf(os.Stderr, "\n\n")
}

// print a message, that the program was terminated because of a detected local deadlock
// Returns:
//  nil
func reportDeadlockPeriodical() {
	fmt.Fprintf(os.Stderr, red, "THE PROGRAM WAS TERMINATED BECAUSE IT DETECTED A LOCAL DEADLOCK\n\n")
}

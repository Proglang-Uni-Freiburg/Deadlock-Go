package deadlock

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
callerInfo.go
Implementation of a struct to save the caller info of locks
*/

// Type to save info about caller.
// A caller is an instance where a lock was created or locked.
type callerInfo struct {
	// name of the file with full path
	file string
	// number of the line, in which the lock is created or locked
	line int
	// true: create, false: lock
	create bool
	// string to save the call stack
	callStacks string
}

// newInfo creates and returns a new callerInfo
//  Args:
//   file (string): name of the file
//   line (int): line in the file where the call happened
//   create (bool): set to true if the call was a lock creation or false, if it was a lock acquiring
//  Returns:
//   callerInfo: the created callerInfo
func newInfo(file string, line int, create bool, callStack string) callerInfo {
	return callerInfo{
		file:       file,
		line:       line,
		create:     create,
		callStacks: callStack,
	}
}

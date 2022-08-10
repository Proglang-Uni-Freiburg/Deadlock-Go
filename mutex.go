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
mutex.go
This file implements the drop-in-replacement for the locks (mutexes) as well as
the lock and unlock operations for these locks.
*/

import (
	"runtime"
	"sync"
	"unsafe"
)

// Type to implement a lock
// It can be used as an drop in replacement
type Mutex struct {
	// mutex for the actual locking
	mu *sync.Mutex
	// info about the creation and lock/unlock of this lock
	context []callerInfo
	// set to true after lock was initialized
	in bool
	// numberLocked stores how often the mutex is currently locked
	numberLocked int
	// index of the routine, which holds the lock
	isLockedRoutineIndex map[int]int
	// lock to prevent multiple concurrent writes to isLockedRoutineIndex
	isLockedRoutineIndexLock *sync.Mutex
	// position of the mutex in memory
	memoryPosition uintptr
}

// create and return a new lock, which can be used as a drop-in replacement for
// sync.Mutex
//  Returns:
//   (*Mutex): the created lock
func NewLock() *Mutex {
	// initialize detector if necessary
	if !initialized {
		initialize()
	}

	m := Mutex{
		mu:                       &sync.Mutex{},
		in:                       true,
		isLockedRoutineIndex:     map[int]int{},
		isLockedRoutineIndexLock: &sync.Mutex{},
	}

	// save the position of the NewLock call
	_, file, line, _ := runtime.Caller(1)
	m.context = append(m.context, newInfo(file, line, true, ""))

	// save the memory position of the mutex
	m.memoryPosition = uintptr(unsafe.Pointer(&m))

	return &m
}

// ============ GETTER ============

// getter for isLocked
//  Returns:
//   (*int): numberLocked
func (m *Mutex) getNumberLocked() *int {
	return &m.numberLocked
}

// getter for isLockedRoutineIndex
//  Returns:
//   (*int): isLockedRoutineIndex
func (m *Mutex) getIsLockedRoutineIndex() *map[int]int {
	return &m.isLockedRoutineIndex
}

// getter for isLockedRoutineIndexLock
//  Returns:
//   (*sync.RWMutex): isLockedRisLockedRoutineIndexLockoutineIndex
func (m *Mutex) getIsLockedRoutineIndexLock() *sync.Mutex {
	return m.isLockedRoutineIndexLock
}

// getter for context
//  Returns:
//   (*[]callerInfo): caller info of the lock
func (m *Mutex) getContext() *[]callerInfo {
	return &m.context
}

// getter for memoryPosition
//  Returns:
//   (uintptr): memoryPosition
func (m *Mutex) getMemoryPosition() uintptr {
	return m.memoryPosition
}

// getter for in
//  Returns:
//   (bool): true if the lock was initialized, false otherwise
func (m *Mutex) getIn() *bool {
	return &m.in
}

// getter for mu
//  Returns:
//   (bool): true, false for rw-mutex
//   (*sync.Mutex): underlying sync.Mutex mu
//   (*sync.RWMutex): nil, underlying sync.RWMutex mu for rw-mutex
func (m *Mutex) getLock() (bool, *sync.Mutex, *sync.RWMutex) {
	return true, m.mu, nil
}

// empty getter, needed for MutexInt
func (m *Mutex) getRLock(routineIndex int) bool {
	return false
}

// empty setter, needed for mutexInt

func (m *Mutex) setRLock(routineIndex int, value bool) {}

// ============ FUNCTIONS ============

// Lock mutex m
//  Returns:
//   nil
func (m *Mutex) Lock() {
	// call the lock function with the mutexInt interface
	lockInt(m, false)
}

// TryLock mutex m
//  Returns:
//   (bool): true if locking was successful, false otherwise
func (m *Mutex) TryLock() bool {
	// call the try-lock method for the mutexInt interface
	return tryLockInt(m, false)
}

// Unlock mutex m
//  Returns:
//   nil
func (m *Mutex) Unlock() {
	if opts.activated {
		// call the unlock method for the mutexInt interface
		unlockInt(m)
	}
	m.mu.Unlock()
}

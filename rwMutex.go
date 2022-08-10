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
rwMutex.go
This file implements the drop-in-replacement for the rw-locks (rw-mutexes) as
well as
the lock, r-lock and unlock operations for these locks.
*/

import (
	"runtime"
	"sync"
	"unsafe"
)

// type to implement a lock
type RWMutex struct {
	// rw-mutex for the actual locking
	mu *sync.RWMutex
	// info about the creation and lock/unlock of this lock
	context []callerInfo
	// set to true after lock was initialized
	in bool
	// how ofter is the lock locked
	numberLocked int
	// indexes of the routines, which holds the lock
	isLockedRoutineIndex map[int]int
	// lock to prevent multiple concurrent writes to isLockedRoutineIndex
	isLockedRoutineIndexLock *sync.Mutex
	// position of the mutex in memory
	memoryPosition uintptr
	// save for the routine index if the lock was locked by rLock
	isRLock map[int]bool
	// lock to prevent concurrent writes to isRLock
	isRLockLock *sync.Mutex
}

// create a new rw-lock
func NewRWLock() *RWMutex {
	// initialize detector if necessary
	if !initialized {
		initialize()
	}

	m := RWMutex{
		mu:                       &sync.RWMutex{},
		in:                       true,
		isLockedRoutineIndex:     map[int]int{},
		isLockedRoutineIndexLock: &sync.Mutex{},
		isRLock:                  map[int]bool{},
		isRLockLock:              &sync.Mutex{},
	}

	// save the position of the NewLock call
	_, file, line, _ := runtime.Caller(1)
	m.context = append(m.context, newInfo(file, line, true, ""))

	// save the memory position of the mutex
	m.memoryPosition = uintptr(unsafe.Pointer(&m))

	return &m
}

// ====== GETTER ===============================================================

// getter for isLocked
//  Returns:
//   (*int): numberLocked
func (m *RWMutex) getNumberLocked() *int {
	return &m.numberLocked
}

// getter for isLockedRoutineIndex
//  Returns:
//   (*int): isLockedRoutineIndex
func (m *RWMutex) getIsLockedRoutineIndex() *map[int]int {
	return &m.isLockedRoutineIndex
}

// getter for isLockedRoutineIndexLock
//  Returns:
//   (*int): isLockedRoutineIndex
func (m *RWMutex) getIsLockedRoutineIndexLock() *sync.Mutex {
	return m.isLockedRoutineIndexLock
}

// getter for context
//  Returns:
//   (*[]callerInfo): caller info of the lock
func (m *RWMutex) getContext() *[]callerInfo {
	return &m.context
}

// getter for memoryPosition
//  Returns:
//   (uintptr): memoryPosition
func (m *RWMutex) getMemoryPosition() uintptr {
	return m.memoryPosition
}

// getter for in
//  Returns:
//   (bool): true if the lock was initialized, false otherwise
func (m *RWMutex) getIn() *bool {
	return &m.in
}

// getter for mu
//  Returns:
//   (bool): false, true for mutex
//   (*sync.Mutex): nil, underlying sync.Mutex mu for mutex
//   (*sync.RWMutex): nil, underlying sync.RWMutex mu
func (m *RWMutex) getLock() (bool, *sync.Mutex, *sync.RWMutex) {
	return false, nil, m.mu
}

// get whether the lock was created by an rlock
//  Args:
//   routineIndex (int): index of the routine
//  Returns:
//   bool. true if it was last locked by rlock, false otherwise
func (m *RWMutex) getRLock(routineIndex int) bool {
	ok, res := m.isRLock[routineIndex]
	if !ok {
		return false
	}
	return res
}

// set whether the lock was created by an rlock
//  Args:
//   routineIndex (int): index of the routine
//   value (bool): true if it was last locked from a rLock, false otherwise
//  Returns:
//   nil
func (m *RWMutex) setRLock(routineIndex int, value bool) {
	m.isLockedRoutineIndexLock.Lock()
	m.isRLock[routineIndex] = value
	m.isLockedRoutineIndexLock.Unlock()
}

// ====== FUNCTIONS ============================================================

// Lock rw-mutex m
//  Returns:
//   nil
func (m *RWMutex) Lock() {
	// call the lock method for the mutexInt interface
	lockInt(m, false)
}

// R-Lock rw-mutex m
//  Returns:
//   nil
func (m *RWMutex) RLock() {
	// call the lock method for the mutexInt interface
	lockInt(m, true)
}

// TryLock rw-mutex m
//  Returns:
//   (bool): true if locking was successful, false otherwise
func (m *RWMutex) TryLock() bool {
	// call the try-lock method for the mutexInt interface
	res := tryLockInt(m, false)
	return res
}

// TryLock rw-mutex m
//  Returns:
//   (bool): true if locking was successful, false otherwise
func (m *RWMutex) RTryLock() bool {
	// call the try-lock method for the mutexInt interface
	res := tryLockInt(m, false)
	return res
}

// Unlock rw-mutex m
//  Returns:
//   nil
func (m *RWMutex) Unlock() {
	if opts.activated {
		unlockInt(m)
	}
	m.mu.Unlock()
}

// Unlock rw-mutex m
//  Returns: nil
func (m *RWMutex) RUnlock() {
	if opts.activated {
		unlockInt(m)
	}
	m.mu.RUnlock()
}

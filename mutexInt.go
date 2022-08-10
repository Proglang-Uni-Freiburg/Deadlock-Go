package deadlock

import (
	"fmt"
	"runtime"
	"sync"
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
mutexInt.go
This file implements and interface for Mutex and RWMutex.
It also implements code which is used for both mutex and rw-mutex
*/

// creat and interface for Mutex and RWMutex
type mutexInt interface {
	// getter for isLocked
	getNumberLocked() *int
	// getter for isLockedRoutineIndex
	getIsLockedRoutineIndex() *map[int]int
	// getter for isLockedRoutineIndexLock
	getIsLockedRoutineIndexLock() *sync.Mutex
	// getter for context
	getContext() *[]callerInfo
	// getter for memoryPosition
	getMemoryPosition() uintptr
	// getter for in (initialized)
	getIn() *bool
	// getter for mu
	// 	if bool is true, *sync.Mutex was returned, *sync.RWMutex is nil
	// 	if bool is false, *sync.Mutex is nil, *sync.RWMutex ware returned
	getLock() (bool, *sync.Mutex, *sync.RWMutex)
	// get whether the lock was created by an rlock
	getRLock(routineIndex int) bool
	// setter for rlock
	setRLock(routineIndex int, value bool)
}

// lock the mutex or rw-mutex and update the detector data
//  Args:
//   m (mutexInt): mutex or rw-mutex to lock
//   rLock (bool): if set to true, the lock is a reader lock
//  Returns:
//   nil
func lockInt(m mutexInt, rLock bool) {
	// do only the operation if detection is completely deactivated
	if !opts.activated {
		d, l, t := m.getLock()
		if d {
			// lock if m is mutex
			l.Lock()
		} else {
			// lock if m is rw-mutex
			if rLock {
				t.RLock()
			} else {
				t.Lock()
			}
		}
		return
	}

	// panic if the lock was not initialized
	if !*m.getIn() {
		errorMessage := fmt.Sprint("Lock ", &m, " was not created. Use ",
			"x := NewLock().")
		panic(errorMessage)
	}

	// defer the actual locking
	defer func() {
		d, l, t := m.getLock()
		if d {
			// lock if m is mutex
			l.Lock()
		} else {
			// lock if m is rw-mutex
			if rLock {
				t.RLock()
			} else {
				t.Lock()
			}
		}

		*m.getNumberLocked() += 1
	}()

	// return if detection is disabled
	if !opts.periodicDetection && !opts.comprehensiveDetection {
		return
	}

	// create new routine, if not initialized
	index := getRoutineIndex()
	if index == -1 {
		newRoutine()
	}
	index = getRoutineIndex()

	r := &routines[index]

	// check if the locking would lead to double locking
	if opts.checkDoubleLocking && *m.getNumberLocked() != 0 {
		r.checkDoubleLocking(m, index, rLock)
	}

	m.getIsLockedRoutineIndexLock().Lock()
	(*m.getIsLockedRoutineIndex())[index] += 1
	m.getIsLockedRoutineIndexLock().Unlock()

	// update data structures if more than on routine is running
	numRoutine := runtime.NumGoroutine()
	if numRoutine > 1 {
		(*r).updateLock(m, rLock)
	}
}

// try to lock the mutex or rw-mutex and update the detector data.
// The lock is only acquired, if it is available at the time of the call
//  Args:
//   m (mutexInt): mutex or rw-mutex to lock
//   rLock (bool): if set to true, the lock is a reader lock
//  Returns:
//   (bool): true if the acquisition was successful, false otherwise
func tryLockInt(m mutexInt, rLock bool) bool {
	// do only the operation if detection is completely deactivated
	if !opts.activated {
		d, l, t := m.getLock()
		var res bool
		if d {
			// lock if m is mutex
			res = l.TryLock()
		} else {
			// lock if m is rw-mutex
			if rLock {
				res = t.TryRLock()
			} else {
				res = t.TryLock()
			}
		}
		return res
	}

	// panic if the lock was not initialized
	if !*m.getIn() {
		errorMessage := fmt.Sprint("Lock ", &m, " was not created. Use ",
			"x := NewLock()")
		panic(errorMessage)
	}

	// try to lock mu
	d, l, t := m.getLock()
	var res bool
	if d {
		// lock if m is mutex
		res = l.TryLock()
	} else {
		// lock if m is rw-mutex
		if rLock {
			res = t.TryRLock()
		} else {
			res = t.TryLock()
		}
	}

	// if locking was successful increase numberLocked
	var index int
	if res {
		// initialize routine if necessary
		index := getRoutineIndex()
		if index == -1 {
			// create new routine, if not initialized
			newRoutine()
		}
		index = getRoutineIndex()

		*m.getNumberLocked() += 1
		m.getIsLockedRoutineIndexLock().Lock()
		(*m.getIsLockedRoutineIndex())[index] += 1
		m.getIsLockedRoutineIndexLock().Unlock()
	}

	// return if detection is disabled
	if !opts.periodicDetection && !opts.comprehensiveDetection {
		return res
	}

	// update data structures if more than on routine is running and locking
	// was successful
	if runtime.NumGoroutine() > 1 {
		if res {
			r := &routines[index]
			(*r).updateTryLock(m, rLock)
		}
	}

	return res
}

// unlock the mutex or rw-mutex and update the detector data
//  Args:
//   m (mutexInt): mutex or RWMutex to unlock
//  Returns:
//   nil
func unlockInt(m mutexInt) {
	// panic if the lock was not initialized
	if !*m.getIn() {
		errorMessage := fmt.Sprint("Lock ", &m, " was not created. Use ",
			"x := NewLock()")
		panic(errorMessage)
	}

	// panic if lock was not locked
	if *m.getNumberLocked() == 0 {
		errorMessage := fmt.Sprint("Tried to unLock lock ", &m,
			" which was not locked.")
		panic(errorMessage)
	}

	// defer the actual unlocking
	defer func() {
		// update numberLocked and isLockedRoutineIndex
		*m.getNumberLocked() -= 1
		m.getIsLockedRoutineIndexLock().Lock()
		(*m.getIsLockedRoutineIndex())[getRoutineIndex()] -= 1
		m.getIsLockedRoutineIndexLock().Unlock()
	}()

	// return if detection is disabled
	if !opts.periodicDetection && !opts.comprehensiveDetection {
		return
	}

	// update data structures if more than on routine is running
	index := getRoutineIndex()
	r := &routines[index]
	(*r).updateUnlock(m)
}

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
routine.go
Implementation of the structure to save the routine wise saved data.
This contains mainly the lock-tree for each routine as well as functionality
to update these trees.
*/

import (
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/petermattis/goid"
)

// map to map the internal routine id to index in routines
var mapIndex = make(map[int64]int)

// lock for the creation of a new routine
var createRoutineLock sync.Mutex

// list of routines
var routines = make([]routine, opts.maxRoutines)

// number of routines in routines
var numberRoutines = 0

// type to implement structures for lock logging
type routine struct {
	// index of the routine
	index int
	// number of currently hold locks
	holdingCount int
	// set of currently hold locks
	holdingSet []mutexInt
	// map of the dependencies
	dependencyMap map[uintptr]*[]*dependency
	// list of dependencies, implements the lock tree
	dependencies [](*dependency)
	// last inserted dependency
	curDep *dependency
	// number of dependencies in dependency map
	depCount int
	// map to save information about collected single level
	collectedSingleLevelLocks map[string][]int
}

// Initialize a go routine
// Returns:
//  nil
func newRoutine() {
	// return if detection is disabled
	if !opts.periodicDetection && !opts.comprehensiveDetection {
		return
	}

	// lock the routine list
	createRoutineLock.Lock()

	// create the routine
	r := routine{
		index:                     numberRoutines,
		holdingCount:              0,
		holdingSet:                make([]mutexInt, opts.maxNumberOfDependentLocks),
		dependencyMap:             make(map[uintptr]*[]*dependency),
		dependencies:              make([]*dependency, opts.maxDependencies),
		curDep:                    nil,
		depCount:                  0,
		collectedSingleLevelLocks: make(map[string][]int),
	}

	// the routine list can only contain a fixed amount of routines
	// panic if it already full
	if numberRoutines >= opts.maxRoutines {
		panic(`Number of routines is greater than max number of routines. 
			Increase Opts.MaxRoutines.`)
	}

	// set the routine
	routines[numberRoutines] = r

	// save the link from internal go id to index of routine
	mapIndex[goid.Get()] = numberRoutines

	// increase number of routines in routine
	numberRoutines++

	// release list lock
	createRoutineLock.Unlock()

	// allocate the dependency list
	// for i := 0; i < opts.maxDependencies; i++ {
	// 	dep := newDependency(nil, nil, 0)
	// 	r.dependencies[i] = &dep
	// }
}

// Update the routine structure if a mutex is locked
// Args:
//  m (mutexInt): mutex to lock
// Returns:
//  nil
func (r *routine) updateLock(m mutexInt, rLock bool) {
	hc := r.holdingCount

	m.setRLock(r.index, rLock)

	isNew := false

	// if lock is not a single level lock -> found nested lock
	if hc > 0 {
		// calculate the key corresponding to the dependency from the memory addresses
		// of m and the last mutex which was added to the list of mutexes which
		// are currently held by r
		key := m.getMemoryPosition() ^ r.holdingSet[hc-1].getMemoryPosition()

		depMap := r.dependencyMap

		// check if the key already exists in depMap
		d, ok := depMap[key]

		panicMassage := `Number of dependencies is greater than max number of 
			dependencies. Increase Opts.MaxDependencies.`

		// Check if the key does not exists or if it exists, that the current
		// dependency, created by locking m is not already in the list of
		// dependencies associated with that key. In this case the dependency
		// will be added to the lock tree
		if !(ok && r.dependencyAlreadyExists(m, d)) {
			// panic if the number of number of dependencies in the lock tree exceeds
			// it maximum
			if r.depCount >= opts.maxDependencies {
				panic(panicMassage)
			}
			// add the new dependency to the lock tree
			dep := newDependency(m, r.holdingSet, hc)
			r.dependencies[r.depCount] = &dep
			dep.update(m, &r.holdingSet, hc)
			r.depCount++

			// add the dependency to the dependencyMap
			if d != nil {
				*d = append(*d, &dep)
			} else {
				d = &[]*dependency{&dep}
			}
			r.dependencyMap[key] = d

			// set the last added dependency pf the tree
			r.curDep = &dep

			isNew = true
		}

	} else {
		// save information on single level locks if enabled in the options
		// to avoid creating the caller info multiple times
		if opts.collectSingleLevelLockStack {
			// get caller information
			_, file, line, _ := runtime.Caller(3)

			// check if a lock of a single level lock was already called in the same file
			if lines, ok := r.collectedSingleLevelLocks[file]; ok {
				isNew = true

				for _, l := range lines {
					// the call (from this file and line) was already called before
					if l == line {
						isNew = false
						break
					}
				}

				if isNew {
					// add new information
					r.collectedSingleLevelLocks[file] = append(
						r.collectedSingleLevelLocks[file], line)
				}
			} else {
				// add new information if no lock was locked from this file before
				isNew = true
				r.collectedSingleLevelLocks[file] = []int{line}
			}
		}
	}

	// save caller information or call stacks if the dependency situation was
	// added for the first time
	if isNew && (hc > 0 || opts.collectSingleLevelLockStack) {
		var file string
		var line int
		var bufStringCleaned string

		// get the call stack if call stack collection is enabled
		if opts.collectCallStack {
			var bufString string
			buf := make([]byte, opts.maxCallStackSize)
			n := runtime.Stack(buf[:], false)
			bufString = string(buf[:n])
			bufStringSplit := strings.Split(bufString, "\n")
			bufStringCleaned = bufStringSplit[0] + "\n"
			for i := 7; i < len(bufStringSplit); i++ {
				bufStringCleaned += bufStringSplit[i] + "\n"
			}
		}

		// get the file and line from which the locking was initiated
		_, file, line, _ = runtime.Caller(3)

		// add the new caller information
		context := m.getContext()
		*context = append(*context, newInfo(file, line, false, bufStringCleaned))
	}

	// panic if the holding depth exceeds its maximum
	if hc >= opts.maxNumberOfDependentLocks {
		panic(`Holding Count is grater than maximum number of dependent locks. 
		Increase Opts.maxNumberOfDependentLocks.`)
	}

	// add the lock to the holding set of the routine
	r.holdingSet[hc] = m
	r.holdingCount++
}

// check if the dependency which results from locking m already exists in list
//  Args:
//   m (mutexInt): mutex which gets locked
//   depList (*([]*dependency)): list to check in
//  Returns:
//   true if dependency already exist
func (r *routine) dependencyAlreadyExists(m mutexInt, depList *([]*dependency)) bool {
	// traverse depList
	for _, d := range *depList {
		hc := r.holdingCount

		// check if dependency with same lock and holding count exists
		if d.mu == m && d.holdingCount == hc {
			// check if the holdingSets in the dependency and the routine are equal
			i := 0
			for d.holdingSet[i] == r.holdingSet[i] && i < hc {
				i++
			}
			if i == hc {
				return true
			}
		}
	}

	return false
}

// update the routine data structure if tryLock is successfully
// this only updates the holding set
//  Args:
//   m (mutexInt): mutex which was locked
//  Returns:
//   nil
func (r *routine) updateTryLock(m mutexInt, rLock bool) {
	// panic if the number of locks in the holding set exceeds its maximum
	hc := r.holdingCount
	if hc >= opts.maxNumberOfDependentLocks {
		panic(`Holding Count is grater than maximum holding depth. Increase 
			Opts.MaxHoldingDepth.`)
	}

	m.setRLock(r.index, rLock)

	// add the lock to the holding set
	r.holdingSet[hc] = m
	r.holdingCount++
}

// Update the routine data structure is a mutex is unlocked
//  Args:
//   m (mutexInt): mutex which was released
//  Returns:
//   nil
func (r *routine) updateUnlock(m mutexInt) {
	// remove m from the holding set of r
	for i := r.holdingCount - 1; i >= 0; i-- {
		if r.holdingSet[i] == m {
			r.holdingSet = append(r.holdingSet[:i], r.holdingSet[i+1:]...)
			r.holdingSet = append(r.holdingSet, nil)
			r.holdingCount--
			break
		}
	}
}

// Get the index of the routine which calls getRoutineIndex in routines
//  Returns:
//   (int): index of the routine in routines which called getRoutineIndex
func getRoutineIndex() int {
	// get an unique internal routine
	// uses "github.com/petermattis/goid"
	id := goid.Get()

	// get the index corresponding to this id
	createRoutineLock.Lock()
	index, ok := mapIndex[id]
	createRoutineLock.Unlock()

	// return -1 if the routine does not exist
	if !ok {
		return -1
	}

	return index
}

// Check if locking mutex m would lead to double locking
//  Args:
//   m (mutexInt): mutex to check for
//   routineIndex (int): index of the routine that tries to lock m
//   rLock (bool): true, if the tested lock acquisition is a r-lock, false otherwise
//  Returns:
//   nil
func (r *routine) checkDoubleLocking(m mutexInt, routineIndex int, rLock bool) {
	// it can only be double locking, if the routine already holds the lock
	if (*m.getIsLockedRoutineIndex())[routineIndex] == 0 {
		return
	}

	// no double locking of two reader
	if rLock && m.getRLock(routineIndex) {
		return
	}

	// report double locking and terminate the program
	reportDeadlockDoubleLocking(m)
	FindPotentialDeadlocks()
	os.Exit(2)
}

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
detector.go
This file contains all the functionality to detect circles in the lock-trees
and therefor actual or potential deadlocks. It implements the periodical
detection during the runtime of the program as well as the comprehensive
detection after the program has finished.
The periodical detection searches for actual deadlocks and can stop the
program if it is in a deadlock situation.
The comprehensive detection should run as soon as the actual program has finished.
It is based on iGoodLock and reports potential deadlocks in the code.
*/

import (
	"fmt"
	"os"
	"runtime"
)

// ================ Comprehensive Detection ================

// FindPotentialDeadlock is the main function to start the comprehensive
// detection of deadlocks. The comprehensive detection uses depth-first search to
// search for loop in the dependency chains of the created lock trees, which are
// represented by the dependency lists of the routines.
// It has to be run at the end of a program to
// detect potential deadlocks in the program. This can be one by calling
// it as a defer statement at the beginning of the main function of the
// program.
//  Returns:
//   nil
func FindPotentialDeadlocks() {
	// check if comprehensive detection is disabled, and if do abort deadlock
	//detection
	if !opts.comprehensiveDetection {
		return
	}

	// only run detector if at least two routines were running during the
	// execution of the program
	if numberRoutines > 1 {
		// abort check if the lock trees contain less than 2 unique dependencies
		if !isNumberDependenciesGreaterEqualTwo() {
			return
		}

		// start the detection of potential deadlocks
		detect()
	}
}

// isNumberDependenciesGreaterEqualTwo counts the number of unique dependencies in
// all and checks if it is greater or equal two lock trees.
// It is not necessary to run comprehensive detection if less then
// two unique dependencies exists.
//  Returns:
//   (bool) : true, if number of unique dependencies is greater or equal than 2,false otherwise
func isNumberDependenciesGreaterEqualTwo() bool {
	// number of already found unique dependencies
	depCount := 0

	// the dependencyString is used to identify a dependency pattern
	var dependencyString string

	// dependencyStrings are saved, so that equal dependencies are not counted twice
	dependencyMap := make(map[string]struct{})

	// parse all routines
	for i := 0; i < numberRoutines; i++ {
		current := routines[i]

		// parse routine i
		for j := 0; j < current.depCount; j++ {
			dep := current.dependencies[j]

			// get the dependency string and store it in dependencySting
			getDependencyString(&dependencyString, dep)

			// check if the dependency string already exists
			if _, ok := dependencyMap[dependencyString]; !ok {
				// new dependency was found
				dependencyMap[dependencyString] = struct{}{}
				depCount++
			}

			// if more than two unique dep have been found return true
			if depCount == 2 {
				return true
			}
		}
	}

	// return false if depCount never reached 2
	return false
}

// getDependencyString calculates the dependency string for a given
// dependency. The string is the concatenation of the on the memory positions
// of mu of the dependency and the locks in the holdingSet of the dependency.
//  Args:
//   str (*string): the dependency string is stored in str
//   dep (*dependency): dependency for which the string gets calculated
//  Returns:
//   nil
func getDependencyString(str *string, dep *dependency) {
	// add the memory position of mu of dep
	*str = fmt.Sprint(dep.mu.getMemoryPosition())

	// add the memory position of the locks in the lockSet of dep
	for i := 0; i < dep.holdingCount; i++ {
		*str += fmt.Sprint(dep.holdingSet[i].getMemoryPosition())
	}
}

// detect runs the detection for loops in the lock trees
//  Returns:
//   nil
func detect() {
	// visiting gets set to index of the routine on which the search for circles is started
	var visiting int

	// A stack is used to represent the currently explored path in the lock trees.
	// A dependency is added to the path by pushing it on top of the stack.
	stack := newDepStack()

	// If a routine has been used as starting routine of a cycle search, all
	// possible paths have already been explored and therefore have no circle.
	// The dependencies in this routine can therefor be ignored for the rest
	// of the search.
	// They can also be temporarily ignored, if a dependency of this routine
	// is already in the path which is currently explored
	isTraversed := make([]bool, numberRoutines)

	// traverse all routines as starting routine for the loop search
	for i := 0; i < numberRoutines; i++ {
		routine := routines[i]

		visiting = i

		// traverse all dependencies of the given routine as starting routine
		// for potential paths
		for j := 0; j < routine.depCount; j++ {
			dep := routine.dependencies[j]
			isTraversed[i] = true

			// push the dependency on the stack as first element of the currently
			// explored path
			stack.push(dep, i)

			// start the depth-first search to find potential circular paths
			dfs(&stack, visiting, &isTraversed)

			// remove dep from the stack
			stack.pop()
		}
	}
}

// dfs runs the recursive depth-first search.
// Only paths which build a valid chain are explored.
// After a new dependency is added to the currently explored path, it is checked,
// if the path forms a circle.
//  Args:
//   stack (*depStack): stack witch represent the currently explored path
//   visiting int: index of the routine of the first element in the currently explored path
//   isTraversed (*([]bool)): list which stores which routines have already been traversed
//    (either as starting routine or as a routine which already has a dep in the current path)
//  Returns:
//   nil
func dfs(stack *depStack, visiting int, isTraversed *([]bool)) {
	// Traverse through all routines to find the potential next step in the path.
	// Routines with index <= visiting have already been used as starting routine
	// and therefore don't have to been considered again.
	for i := visiting + 1; i < numberRoutines; i++ {
		routine := routines[i]

		// continue if the routine has already been traversed
		if (*isTraversed)[i] {
			continue
		}

		// go through all dependencies of the current routine
		for j := 0; j < routine.depCount; j++ {
			dep := routine.dependencies[j]
			// check if adding dep to the stack would still be a valid path
			if isChain(stack, dep, i) {
				// check if adding dep to the stack would lead to a cycle
				if isCycleChain(stack, dep, i) {
					// report the found potential deadlock
					stack.push(dep, j)
					reportDeadlock(stack)
					stack.pop()
				} else { // the path is not a cycle yet
					// add dep to the current path
					stack.push(dep, i)
					(*isTraversed)[i] = true

					// call dfs recursively to traverse the path further
					dfs(stack, visiting, isTraversed)

					// dep did not lead to a cycle in the lock trees.
					// It is removed to explore different paths
					stack.pop()
					(*isTraversed)[i] = false
				}
			}
		}
	}
}

// ================ Periodical Detection ================

// periodicalDetection is the main function to start the periodical detection.
// It is called periodically to detect if the program is in a local deadlock
// state i.e. a state in which only a subset of the running routines are in
// a deadlock position.
//  The program will be terminated if such a situation occurs and the comprehensive
//  detection will automatically be started.
//  If the program is in a total deadlock, i.e. no routine is running anymore,
//  it is normally automatically terminated my the go-runtime deadlock detection.
//  In this case the comprehensive detection can not be started.
// To detect such local deadlocks, the detector uses the dependency of each
// routine, which was last added to the routine and searches for circles in
// this set of dependencies.
//  Args:
//   lastHolding (*[]mutexInt): list of the dependencies which were considered
//    in the last run
//  Returns:
//   nil
func periodicalDetection(lastHolding *[]mutexInt) {
	// only check if at least two routines are currently running
	if runtime.NumGoroutine() < 2 {
		return
	}

	// A stack is used to represent the currently explored path in the lock trees.
	// A dependency is added to the path by pushing it on top of the stack.

	// the detection is only run if the number of routines which hold at least two
	// locks is at least 2 and the situation has changed since the late periodical check
	nrThreadsHoldingLocks := 0
	sthNew := false

	// traverse all routines
	for index, r := range routines {
		// check if the routine holds at least two lock and the last added dependency
		// has changed since the last check
		holds := r.holdingCount - 1
		if holds >= 0 && (*lastHolding)[index] != r.holdingSet[holds] {
			(*lastHolding)[index] = r.holdingSet[holds]
			sthNew = true
			if holds > 0 {
				nrThreadsHoldingLocks++
			}
		} else if holds < 0 && (*lastHolding)[index] != nil {
			(*lastHolding)[index] = nil
			sthNew = true
		}
	}

	// abort the detection if nothing has changed or not enough routines hold locks
	if !sthNew || nrThreadsHoldingLocks <= 1 {
		return
	}

	// run the detection
	detectionPeriodical(lastHolding)
}

// detectPeriodical starts the search for local deadlocks.
// It uses depth-first search to search for cyclic chains in the set of
// dependencies which contain the dependencies which were last added to each
// routine
// 	Args:
//   lastHolding (*[]mutexInt): list with dependencies
//  Returns:
//   nil
func detectionPeriodical(lastHolding *[]mutexInt) {
	// A stack is used to represent the currently explored path in the lock trees.
	// A dependency is added to the path by pushing it on top of the stack.
	stack := newDepStack()

	// every dependency can only be used once in the path
	isTraversed := make([]bool, opts.maxRoutines)

	// traverse all routines as starting routine
	for index, r := range routines {
		// routines with an index >= routinesIndex have not been used in the program
		if index >= numberRoutines {
			break
		}

		// continue if the routine has not acquired a dependency
		if r.curDep == nil {
			continue
		}

		isTraversed[index] = true

		// add the dependency as first dependency of the path to the stack and
		// start the recursive search for a cyclic path
		stack.push(r.curDep, index)
		dfsPeriodical(&stack, index, isTraversed, lastHolding)

		// if no cycle is found with this dependency it is removed from the path
		stack.pop()
		r.curDep = nil
	}
}

// dfsPeriodical runs the recursive depth-first search.
// Only paths which build a valid chain are explored.
// After a new dependency is added to the currently explored path, it is checked,
// if the path forms a circle.
//  Args:
//   stack (*depStack): stack witch represent the currently explored path
//   visiting int: index of the routine of the first element in the currently explored path
//   isTraversed (*([]bool)): list which stores which routines have already been traversed
//    (either as starting routine or as a routine which already has a dep in the current path)
//   lastHolding (*[]mutexInt): list with dependencies
//  Returns:
//   nil
func dfsPeriodical(stack *depStack, visiting int, isTraversed []bool,
	lastHolding *[]mutexInt) {
	// Traverse through all routines to find the potential next step in the path.
	// Routines with index <= visiting have already been used as starting routine
	// and therefore don't have to been considered again.
	for i := visiting + 1; i < numberRoutines; i++ {
		r := routines[i]

		// continue if the routine has no current dependency or has already be traversed
		if r.curDep == nil || isTraversed[i] {
			continue
		}

		dep := r.curDep

		// check if adding dep to the current path would lead to a valid dependency
		// chain
		if !isChain(stack, dep, i) {
			continue
		}

		// check if adding dep to the curring path would lead to a cyclic dependency
		// chain. This would indicate a deadlock.
		if isCycleChain(stack, dep, i) {
			stack.push(dep, i)

			// check if the last added dependency in on of the routines in the path
			// has changed since the beginning of the detection. In this case, the
			// program will assume it was a false alarm and will not terminate the
			// program
			sthNew := false

			// traverse alle routines in the current dependency chain
			for cl := stack.stack.next; cl != nil; cl = cl.next {
				routineInChain := routines[cl.index]

				// check if the last added dependency has changed
				holds := routineInChain.holdingCount - 1
				if (holds >= 0 &&
					(*lastHolding)[cl.index] != routineInChain.holdingSet[holds]) ||
					(holds < 0 && (*lastHolding)[cl.index] != nil) {
					sthNew = true
					break
				}
			}

			// if nothing has changed the program assumes a deadlock.
			// Therefore it reports the deadlock, starts the comprehensive detection
			// to search for other possible deadlocks and terminates the program.
			if !sthNew {
				reportDeadlockPeriodical()
				FindPotentialDeadlocks()
				os.Exit(2)
			}
			stack.pop()
		} else {
			// if the chain is not a cycle, the dependency is added to the current
			// path and the search is continued recursively
			isTraversed[numberRoutines] = true
			stack.push(dep, numberRoutines)
			dfsPeriodical(stack, visiting, isTraversed, lastHolding)

			// if no cycle has been found with dep, it is removed from the path
			stack.pop()
			isTraversed[numberRoutines] = false
		}
	}
}

// ================ Checks for chains and Cycles ================

// isCain checks if adding dep to the current path represented by stack is
// still a valid path.
//  A valid path contains the same dependency only once and contains the same
//  lock only once. A path is also not valid if there exist two locks in the
// holdings sets of two different dependencies in the path, such that the locks
// are equal. This would be a gate lock situation. For RW-Locks this is not
// true if both of the locks were acquired with RLock, because RLocks don't
// have to work as gate locks
//  Args:
//   stack (*depStack): stack representing the current path
//   dep (*dependency): dependency for which it should be checked if it can be
//    added to the path
//   routineIndex (int): index of the routine the dependency is from
//  Returns:
//   (bool): true if dep can be added to the current path, false otherwise
func isChain(stack *depStack, dep *dependency, routineIndex int) bool {
	// the mutex of the depEntry at the top of the stack mut be in the
	// holding set of dep
	found := false
	for i := 0; i < dep.holdingCount; i++ {
		mutexInHs := dep.holdingSet[i]
		if mutexHaveEqualLock(mutexInHs, stack.top.depEntry.mu) {
			// if mutexInHs is read, the mutex at the top of the stack can not also be read
			if !(mutexInHs.getRLock(routineIndex) && stack.top.depEntry.mu.getRLock(stack.top.index)) {
				found = true
				break
			}
		}
	}
	if !found {
		return false
	}

	for c := stack.stack.next; c != nil; c = c.next {
		// no two dependencies in the stack can be equal
		if c.depEntry == dep {
			return false
		}

		// If two holding sets contain the same mutex they both have to be rLock
		// (gate lock)
		for i := 0; i < dep.holdingCount; i++ {
			for j := 0; j < c.depEntry.holdingCount; j++ {
				lockInDepHs := dep.holdingSet[i]
				lockInCHoldingSet := c.depEntry.holdingSet[j]
				if mutexHaveEqualLock(lockInDepHs, lockInCHoldingSet) {
					if !(lockInCHoldingSet.getRLock(c.index) && lockInDepHs.getRLock(routineIndex)) {
						return false
					}
				}
			}
		}
	}

	return true
}

// isCycleCain checks if adding a dependency dep to the current path represented
// by stack would lead to a cyclic chain, meaning the lock mu of dep is in the
// holding set of the first dependency in the path. This would indicate a possible
// deadlock situation. With RW-locks it is possible, that a cyclic path
// does not indicate a potential deadlock. In this case, the function assumes,
// that the path does not create a valid cyclic chain.
//  isCycleChain assumes, that adding dep to the path results to a valid path
//  (see isChain)
// Args:
//  stack (*depStack): stack representing the current path
//  dep (*dependency): dependency for which it should be checked if adding dep
//   to the path would lead to a cyclic path
//  routineIndex (int): index of the routine from which dep originated
// Returns:
//  (bool): true if dep can be added to the current path to create a valid cyclic
//   chain, false if the path is no cycle, or it contains RW-lock with which
//   the cycle does not indicate a deadlock
func isCycleChain(dStack *depStack, dep *dependency, routineIndex int) bool {
	// the mutex dep must be in the holding set of the depEntry at the bottom of
	// the stack
	found := false
	for i := 0; i < dStack.stack.next.depEntry.holdingCount; i++ {
		mutexInHs := dStack.stack.next.depEntry.holdingSet[i]
		if mutexHaveEqualLock(mutexInHs, dep.mu) {
			// if mutexInHs is read, the mutex at the top of the stack can not also be read
			if !(mutexInHs.getRLock(dStack.stack.index) && dep.mu.getRLock(routineIndex)) {
				found = true
				break
			}
		}
	}
	return found
}

func mutexHaveEqualLock(m1, m2 mutexInt) bool {
	isMutex1, mutex1, rwmutex1 := m1.getLock()
	isMutex2, mutex2, rwmutex2 := m2.getLock()

	if isMutex1 != isMutex2 {
		return false
	}

	if isMutex1 && (mutex1 == mutex2) {
		return true
	}

	if !isMutex1 && (rwmutex1 == rwmutex2) {
		return true
	}

	return false
}

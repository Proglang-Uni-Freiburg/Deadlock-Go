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
dependency.go
Implement structure to describe dependencies.
Each dependency contains the lock l as well as all the locks l depends on.
These are all the locks which were hold by the same routine, when l was
acquired.
*/

// Type to implement a dependency
// A dependency represents a set of edges in a lock tree
// It consist of a lock l and a list of all locks, on which l depends
// i.e. all lock which were already locked by the same routine, when
// l was acquired.
type dependency struct {
	mu           mutexInt   // lock
	holdingSet   []mutexInt // locks which where locked while mu was acquired
	holdingCount int        // on how many locks does mu depend
}

// newDependency creates and returns a new dependency object
//  Args:
//   mu (mutexInt): lock of the dependency
//   currentLocks ([]mutexInt): list of locks mu depends on
//   numberOfLocks (int): number of locks lock depends on
//  Returns:
//   (dependency) : the created dependency
func newDependency(lock mutexInt, currentLocks []mutexInt,
	numberOfLocks int) dependency {
	// create dependency
	d := dependency{
		mu:           lock,
		holdingCount: numberOfLocks,
		holdingSet:   make([]mutexInt, opts.maxNumberOfDependentLocks),
	}

	// copy currentLocks into d.holding set
	for i := 0; i < numberOfLocks; i++ {
		d.holdingSet = append(d.holdingSet, currentLocks[i])
	}

	return d
}

// update updates a dependency
//  Args:
//   lock (mutexInt): new lock of the dependency
//   hs (*[]mutexInt): new holding set
//   numberOfLocks (int): new number of locks lock depends on
//  Returns:
//   nil
func (d *dependency) update(lock mutexInt, hs *[]mutexInt, numberOfLocks int) {
	// set new lock
	d.mu = lock

	// copy hs into the holding set
	copy(d.holdingSet, *hs)

	// set new holdingCount
	d.holdingCount = numberOfLocks
}

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
options.go
This file implements options for the deadlock detections such as the
enabling or disabling of the periodical and/or comprehensive detection as
well as the periodical detection time and max values for the detection.
*/

import "time"

// opts controls how the detection behaves
var opts = struct {
	// if deactivated is false, there is no detection
	activated bool
	// If periodicDetection is set to false, periodic detection is disabled
	periodicDetection bool
	// If comprehensiveDetection is set to false, comprehensive detection at
	// the end of the program is disabled
	comprehensiveDetection bool
	// Set how often the periodic detection is run
	periodicDetectionTime time.Duration
	// If collectCallStack is true, the CallStack for lock creation and
	// acquisition are collected and displayed. Otherwise only file names and
	// lines are collected
	collectCallStack bool
	// If collectSingleLevelLockStack is set to true, stack traces for single
	// level locks are collected. Otherwise not.
	collectSingleLevelLockStack bool
	// If checkDoubleLocking is set to true, the detector checks for double
	// locking
	checkDoubleLocking bool
	// maximum number of dependencies
	maxDependencies int
	// The maximum number of locks a lock can depend on
	maxNumberOfDependentLocks int
	// The maximum number of routines
	maxRoutines int
	// The maximum byte size for callStacks
	maxCallStackSize int
}{
	activated:                   true,
	periodicDetection:           true,
	comprehensiveDetection:      true,
	periodicDetectionTime:       time.Second * 2,
	collectCallStack:            false,
	collectSingleLevelLockStack: true,
	checkDoubleLocking:          true,
	maxDependencies:             4096,
	maxNumberOfDependentLocks:   128,
	maxRoutines:                 1024,
	maxCallStackSize:            2048,
}

// Enable or disable all detections
// It is not possible to set options after the detector was initialized
//  Args:
//   enable (bool): true to enable, false to disable
//  Returns:
//   (bool): true, if the set was successful, false otherwise
func SetActivated(enable bool) bool {
	if initialized {
		return false
	}
	opts.activated = enable
	opts.checkDoubleLocking = true
	opts.periodicDetection = true
	opts.comprehensiveDetection = true
	return true
}

// Enable or disable periodic detection
// It is not possible to set options after the detector was initialized
//  Args:
//   enable (bool): true to enable, false to disable
//  Returns:
//   (bool): true, if the set was successful, false otherwise
func SetPeriodicDetection(enable bool) bool {
	if initialized {
		return false
	}
	opts.periodicDetection = enable
	setActivatedAuto()
	return true
}

// Enable or disable comprehensive detection
// It is not possible to set options after the detector was initialized
//  Args:
//   enable (bool): true to enable, false to disable
//  Returns:
//   (bool): true, if the set was successful, false otherwise
func SetComprehensiveDetection(enable bool) bool {
	if initialized {
		return false
	}
	opts.comprehensiveDetection = enable
	setActivatedAuto()
	return true
}

// Set the temporal distance between the periodic detections
// It is not possible to set options after the detector was initialized
//  Args:
//   seconds (int): temporal distance in seconds
//  Returns:
//   (bool): true, if the set was successful, false otherwise
func SetPeriodicDetectionTime(seconds int) bool {
	if initialized {
		return false
	}
	opts.periodicDetectionTime = time.Second * time.Duration(seconds)
	return true
}

// Enable or disable collection of full call stacks
// If it is disabled only file and line numbers are collected
// It is not possible to set options after the detector was initialized
//  Args:
//   enable (bool): true to enable, false to disable
//  Returns:
//   (bool): true, if the set was successful, false otherwise
func SetCollectCallStack(enable bool) bool {
	if initialized {
		return false
	}
	opts.collectCallStack = enable
	return true
}

// Enable or disable collection of call information for single level locks
// If it is disabled no caller information about single level locks will be collected.
// It is not possible to set options after the detector was initialized
//  Args:
//   enable (bool): true to enable, false to disable
//  Returns:
//   (bool): true, if the set was successful, false otherwise
func SetCollectSingleLevelLockInformation(enable bool) bool {
	if initialized {
		return false
	}
	opts.collectSingleLevelLockStack = enable
	return true
}

// Enable or disable checks for double locking
// It is not possible to set options after the detector was initialized
//  Args:
//   enable (bool): true to enable, false to disable
//  Returns:
//   (bool): true, if the set was successful, false otherwise
func SetDoubleLockingDetection(enable bool) bool {
	if initialized {
		return false
	}
	opts.checkDoubleLocking = enable
	setActivatedAuto()
	return true
}

// Set the max number of dependencies
// It is not possible to set options after the detector was initialized
//  Args:
//   number (int): max number of dependencies
//  Returns:
//   (bool): true, if the set was successful, false otherwise
func SetMaxDependencies(number int) bool {
	if initialized {
		return false
	}
	opts.maxDependencies = number
	return true
}

// Set the max number of locks a lock can depend on
// It is not possible to set options after the detector was initialized
//  Args:
//   number (int): max number of locks a lock can depend on
//  Returns:
//   (bool): true, if the set was successful, false otherwise
func SetMaxNumberOfDependentLocks(number int) bool {
	if initialized {
		return false
	}
	opts.maxNumberOfDependentLocks = number
	return true
}

// Set the max number of routines
// It is not possible to set options after the detector was initialized
//  Args:
//   number (int): max number of routines
//  Returns:
//   (bool): true, if the set was successful, false otherwise
func SetMaxRoutines(number int) bool {
	if initialized {
		return false
	}
	opts.maxRoutines = number
	return true
}

// Set the max size of collected call stacks
// It is not possible to set options after the detector was initialized
//  Args:
//   number (int): max size of the call stack in bytes
//  Returns:
//   (bool): true, if the set was successful, false otherwise
func SetMaxCallStackSize(number int) bool {
	if initialized {
		return false
	}
	opts.maxCallStackSize = number
	return true
}

// automatically set activated according to the other options
//  Returns:
//   nil
func setActivatedAuto() {
	if !(opts.periodicDetection || opts.checkDoubleLocking || opts.comprehensiveDetection) {
		opts.activated = false
		return
	}
	opts.activated = true

}

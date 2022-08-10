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
initialize.go
This code initializes the deadlock detector. Its main task is to start
and periodically run the periodical deadlock detection.
*/

import (
	"time"
)

// global variable to check whether the detector was already initialized
var initialized = false

// initialize initializes the deadlock detector.
// This starts the periodical detection.
//  Returns:
//   nil
func initialize() {
	initialized = true

	// reinitialize routines to set size
	routines = make([]routine, opts.maxRoutines)

	// return if periodical detection is disabled
	if !opts.periodicDetection {
		return
	}

	// go routine to run the periodical detection in the background
	go func() {
		// timer to send a signals at equal intervals
		timer := time.NewTicker(opts.periodicDetectionTime)

		// initialize lashHolding. This slice stores the dependencies which were
		// considered in the last detection round, so that the detection only takes
		// place, if the situation has changed
		lastHolding := make([]mutexInt, opts.maxRoutines)

		// run the periodical detection if a timer signal is received
		for range timer.C {
			periodicalDetection(&lastHolding)
		}
	}()
}

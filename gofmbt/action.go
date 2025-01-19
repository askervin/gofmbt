// Copyright Antti Kervinen. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you
// may not use this file except in compliance with the License.  You
// may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.  See the License for the specific language governing
// permissions and limitations under the License.

package gofmbt

import (
	"fmt"
)

// Action, associated with a state change, specifies what to execute
// to cause corresponding state change on the system under test.
type Action struct {
	name   string
	format string
	args   []interface{}
}

// NewAction creates a new action.
func NewAction(format string, args ...interface{}) *Action {
	return &Action{
		format: format,
		args:   args,
		name:   fmt.Sprintf(format, args...),
	}
}

// String returns a string representation of an action.
func (a *Action) String() string {
	return a.name
}

// When returns a slice containing transitions if enabled is
// true. This is a convenience function for When/OnAction/Do modeling
// syntax.
func When(enabled bool, tss ...[]*Transition) []*Transition {
	if !enabled {
		return nil
	}
	ts := []*Transition{}
	for _, origTs := range tss {
		ts = append(ts, origTs...)
	}
	return ts
}

// OnAction returns new Action. This is a convenience function for
// When/OnAction/Do modeling syntax.
func OnAction(format string, args ...interface{}) *Action {
	return NewAction(format, args...)
}

// Do returns a slice containing one transition. Do is a convenience
// function for When/OnAction/Do modeling syntax.
func (a *Action) Do(stateChanges ...StateChange) []*Transition {
	stateChange := func(s State) State {
		for _, sc := range stateChanges {
			s = sc(s)
		}
		return s
	}
	return []*Transition{NewTransition(a, stateChange)}
}

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

// StateChange is a function associated to a transition. The function
// returns new state that reflects the change in the state space.
// StateChange function must not modify the original state.  If
// StateChange returns nil, the transition is not possible in the
// original state.
type StateChange func(State) State

// TransitionGen is a function that generates transitions from a given
// state.
type TransitionGen func(State) []*Transition

// Transition represents state changes by an action in various
// states.
type Transition struct {
	action      *Action
	stateChange StateChange
}

// NewTransition creates a new transition.
func NewTransition(a *Action, sc StateChange) *Transition {
	return &Transition{a, sc}
}

// Action returns the action of a transition.
func (t *Transition) Action() *Action {
	return t.action
}

// StateChange calls the state change function of a transition. It
// returns the new state after the transition or nil if the transition
// is not possible in the original state.
func (t *Transition) StateChange(s State) State {
	return t.stateChange(s)
}

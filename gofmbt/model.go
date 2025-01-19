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

// Walkable models can be explored: which steps can be taken next,
// what paths are possible. Yet paths can be constructed through
// repeated "which steps" calls, they are included in the interface
// separately to enable selective caching. For instance, path cache
// can be implemented as a walkable object that caches Paths calls
// directly without a step-by-step state cache.
type Walkable interface {
	// StepsFrom returns all alternative steps that start from a
	// given state.
	StepsFrom(s State) []*Step
	// Paths returns all alternative paths of at most maxLen steps
	// that start from a given state.
	Paths(s State, maxLen int) []Path
}

// Model specifies a state space.
type Model struct {
	gen []TransitionGen // transition generators
}

// NewModel creates a new model.
func NewModel() *Model {
	return &Model{}
}

// From adds a transition generator to the model.
func (m *Model) From(transitionGen func(State) []*Transition) {
	m.gen = append(m.gen, transitionGen)
}

// TransitionsFrom returns all transitions that may be taken from a given state.
func (m *Model) TransitionsFrom(s State) []*Transition {
	ts := []*Transition{}
	for _, gen := range m.gen {
		ts = append(ts, gen(s)...)
	}
	return ts
}

// Steps returns all steps that start from a given state.
func (m *Model) StepsFrom(s State) []*Step {
	steps := []*Step{}
	for _, t := range m.TransitionsFrom(s) {
		if endState := t.stateChange(s); endState != nil {
			steps = append(steps, NewStep(s, t.action, endState))
		}
	}
	return steps
}

// Paths returns all alternative paths of at most maxLen steps
// that start from a given state.
func (m *Model) Paths(s State, maxLen int) []Path {
	paths := []Path{}
	if maxLen == 0 {
		return nil
	}
	for _, step := range m.StepsFrom(s) {
		newPath := NewPath(step)
		childPaths := m.Paths(step.EndState(), maxLen-1)
		if len(childPaths) == 0 {
			paths = append(paths, newPath)
		} else {
			for _, childPath := range childPaths {
				paths = append(paths, append(newPath, childPath...))
			}
		}
	}
	if len(paths) == 0 {
		return nil
	}
	return paths
}

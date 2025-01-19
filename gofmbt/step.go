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

// Step represents a step in a path.
type Step struct {
	start  State
	action *Action
	end    State
}

// Path represents a path in a model.
type Path []*Step

// NewStep creates a new step.
func NewStep(start State, action *Action, end State) *Step {
	return &Step{start, action, end}
}

// String returns a string representation of a step.
func (step *Step) String() string {
	return fmt.Sprintf("[%s--%s->%s]", step.start, step.action, step.end)
}

// StartState returns the start state of a step.
func (step *Step) StartState() State {
	return step.start
}

// Action returns the action of a step.
func (step *Step) Action() *Action {
	return step.action
}

// EndState returns the end state of a step.
func (step *Step) EndState() State {
	return step.end
}

// NewPath creates a new path from steps.
func NewPath(steps ...*Step) Path {
	path := Path{}
	for _, step := range steps {
		path = append(path, step)
	}
	return path
}

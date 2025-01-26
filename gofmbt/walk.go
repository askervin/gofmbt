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
	"iter"
)

type StepFilter func([]*Step) []*Step

type Walker struct {
	m          Walkable
	stepFilter StepFilter
}

func NewWalker(m Walkable) *Walker {
	return &Walker{m: m}
}

func (w *Walker) SetStepFilter(f StepFilter) {
	w.stepFilter = f
}

// IterPaths yields all possible paths of at most maxLen steps
// starting from a state. Note: paths are yielded as slices of the
// same list. Therefore a yielded path must be copied in order to save
// it.
func (w *Walker) IterPaths(s State, maxLen int) iter.Seq[Path] {
	path := make(Path, maxLen, maxLen)
	return func(yield func(Path) bool) {
		w.yieldPaths(yield, &path, 0, s, maxLen)
	}
}

func (w *Walker) yieldPaths(yield func(Path) bool, path *Path, index int, s State, maxLen int) bool {
	if index == maxLen {
		return yield(*path)
	}
	nextSteps := w.m.StepsFrom(s)
	if w.stepFilter != nil {
		nextSteps = w.stepFilter(nextSteps)
	}
	if len(nextSteps) == 0 {
		return yield((*path)[:index])
	}
	for _, step := range nextSteps {
		(*path)[index] = step
		if !w.yieldPaths(yield, path, index+1, step.EndState(), maxLen) {
			return false
		}
	}
	return true
}

// Paths returns all alternative paths of at most maxLen steps
// that start from a given state.
func (w *Walker) Paths(s State, maxLen int) []Path {
	paths := []Path{}
	for path := range w.IterPaths(s, maxLen) {
		pathCopy := make(Path, len(path))
		copy(pathCopy, path)
		paths = append(paths, pathCopy)
	}
	return paths
}

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
	"testing"
)

type PlayerState struct {
	playing bool
	song    int
}

func (ps *PlayerState) String() string {
	return fmt.Sprintf("{playing:%v,song:%v}", ps.playing, ps.song)
}

// This is Rosetta's stone on modelling directly with StateChange
// functions and NewTransitions in newPlayerModelWithRawTransitions,
// and writing exactly the same model using When/OnAction/Do.
func newPlayerModelWithRawTransitions() *Model {
	model := NewModel()
	play := func(current State) State {
		s := current.(*PlayerState)
		if s.playing {
			return nil
		}
		return &PlayerState{
			playing: true,
			song:    s.song,
		}
	}
	pause := func(current State) State {
		s := current.(*PlayerState)
		if !s.playing {
			return nil
		}
		return &PlayerState{
			playing: false,
			song:    s.song,
		}
	}
	nextsong := func(current State) State {
		s := current.(*PlayerState)
		if s.song >= 3 {
			return nil
		}
		return &PlayerState{
			playing: s.playing,
			song:    s.song + 1,
		}
	}
	prevsong := func(current State) State {
		s := current.(*PlayerState)
		if s.song <= 1 {
			return nil
		}
		return &PlayerState{
			playing: s.playing,
			song:    s.song - 1,
		}
	}

	model.From(func(current State) []*Transition {
		return []*Transition{
			NewTransition(NewAction("play"), play),
			NewTransition(NewAction("pause"), pause),
			NewTransition(NewAction("nextsong"), nextsong),
			NewTransition(NewAction("prevsong"), prevsong),
		}
	})

	return model
}

func newPlayerModelWithWhenOnAction() *Model {
	setState := func(playing bool, song int) StateChange {
		return func(_ State) State {
			return &PlayerState{playing, song}
		}
	}

	model := NewModel()
	model.From(func(start State) []*Transition {
		s := start.(*PlayerState)
		return When(true,
			When(s.playing,
				OnAction("pause").Do(setState(false, s.song))),
			When(!s.playing,
				OnAction("play").Do(setState(true, s.song))),
			When(s.song < 3,
				OnAction("nextsong").Do(setState(s.playing, s.song+1))),
			When(s.song > 1,
				OnAction("prevsong").Do(setState(s.playing, s.song-1))),
		)
	})
	return model
}

var playerModels map[string]*Model = map[string]*Model{
	"raw":  newPlayerModelWithRawTransitions(),
	"when": newPlayerModelWithWhenOnAction(),
}

func TestCoverPlayerActions(t *testing.T) {
	for modelName, model := range playerModels {
		state := &PlayerState{false, 1}
		coverer := NewCoverer()
		coverer.CoverActionCombinations(1)
		path, stats := coverer.BestPath(model, state, 6)
		if len(path) != 6 {
			t.Fatalf("model %q: expected len(path)==6, got %d", modelName, len(path))
		}
		if stats.MaxIncrease != 4 {
			t.Fatalf("model %q: expected reaching coverage 4, got %d", modelName, stats.MaxIncrease)
		}
		if stats.MaxStep != 3 {
			t.Fatalf("model %q: expected all actions covered at step 4, got %d", modelName, stats.MaxStep)
		}

		covered := map[string]int{}
		for _, step := range path[:stats.MaxStep+1] {
			covered[step.Action().String()] += 1
		}
		if len(covered) != 4 {
			t.Fatalf("model %q: expected 4 different actions, got %d", modelName, len(covered))
		}
		for _, expectedStep := range []string{"play", "pause", "nextsong", "prevsong"} {
			if _, ok := covered[expectedStep]; !ok {
				t.Fatalf("model %q: expected %q in covered, got: %v", modelName, expectedStep, covered)
			}
		}
	}
}

func TestCoverPlayerStates(t *testing.T) {
	for modelName, model := range playerModels {
		state := &PlayerState{false, 1}
		coverer := NewCoverer()
		coverer.CoverStates()
		path, stats := coverer.BestPath(model, state, 8)
		if len(path) != 8 {
			t.Fatalf("model %q: expected len(path)==10, got %d", modelName, len(path))
		}
		if stats.MaxIncrease != 6 {
			t.Fatalf("model %q: expected reaching 6 different states, got %d", modelName, stats.MaxIncrease)
		}
		if stats.MaxStep != 4 {
			t.Fatalf("model %q: expected all stats visited at step 4, got %d", modelName, stats.MaxStep)
		}
		if len(coverer.CoveredStrings()) != 0 {
			t.Fatalf("model %q: expected nothing to be covered yet, got %v", modelName, coverer.CoveredStrings())
		}
		coverer.MarkCovered(path[:stats.MaxStep+1]...)
		coverer.UpdateCoverage()
		if len(coverer.CoveredStrings()) != 6 {
			t.Fatalf("model %q: expected 6 states covered, got %d %v", modelName, len(coverer.CoveredStrings()), coverer.CoveredStrings())
		}
	}
}

func TestStepsFrom(t *testing.T) {
	for modelName, model := range playerModels {
		state := &PlayerState{false, 1}
		altSteps := model.StepsFrom(state)
		if len(altSteps) != 2 {
			t.Fatalf("model %q: expected 2 StepsFrom %s, got %d", modelName, state, len(altSteps))
		}
		for _, step := range altSteps {
			if step.Action().String() == "play" {
				continue
			}
			if step.Action().String() == "nextsong" {
				continue
			}
			t.Fatalf("model %q: unexpected action %q in state %s. Step: %s", modelName, step.Action(), state, step)
		}
	}
}

func TestSearchPathsTestSteps(t *testing.T) {
	for modelName, model := range playerModels {
		state := &PlayerState{false, 1}
		coverer := NewCoverer()
		coverer.CoverStateActions()
		t.Log(modelName)
		i := 0
		for {
			path, stats := coverer.BestPath(model, state, 6)
			if len(path) == 0 {
				break
			}
			// Verify prev step.EndState() == next step.StartState()
			var prev *Step
			for _, step := range path {
				if prev != nil && step.StartState().String() != prev.EndState().String() {
					t.Fatalf("model %q: prev end state and next start state differ: %s != %s", modelName, prev, step)
				}
				prev = step
			}

			// Execute path up to the first step where
			// coverage increases and search for the
			// BestPath continuing from there.
			for _, step := range path[:stats.FirstStep+1] {
				i++
				t.Log("step:", step)
				coverer.MarkCovered(step)
				coverer.UpdateCoverage()
				t.Log("coverage:", coverer.Coverage(), "strings:", coverer.CoveredStrings())
			}
			state = path[stats.FirstStep].EndState().(*PlayerState)
		}
		if coverer.Coverage() != 14 {
			t.Fatalf("model %q: expected 14 {state}--action--> combinations to be covered, got %d", modelName, coverer.Coverage())
		}
		if i != 14 {
			t.Fatalf("model %q: expected reach max coverage with 14 steps, needed: %d", modelName, i)
		}
	}
}

func BenchmarkPathsSmallStateSpace(b *testing.B) {
	for modelName, model := range playerModels {
		state := &PlayerState{false, 1}
		depth := 12
		paths := model.Paths(state, depth)
		b.Log("found", len(paths), "paths of depth", depth, "from", modelName)
	}
}

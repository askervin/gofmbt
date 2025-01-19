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

// Package gofmbt implements a model-based testing library, including
// tools for
//   1. defining test models
//   2. defining what a test should cover
//   3. generating tests sequences with optimal coverage
//
// # Models
//
// A model specifies what and when can be tested. "What" is specified
// by an Action while "when" is expressed using states and
// transitions.
//
// State is implemented by user. It needs to implement String(). For
// example, if testing a music player, a State can be defined as:
//
//  type PlayerState struct {
//         playing bool // player is either playing or paused
//         song    int  // the number of the song being played
//  }
//
// StateChange is a function that takes a start state as an input and
// returns an end state as an output. If a StateChange function
// returns nil, the state change is unspecified at the start
// state. User defines typically many StateChange functions, each of
// which modify one or more attributes of State. For example, a state
// change that starts a paused player, but is unspecified if the
// player is already playing, can be defined as:
//
//  func startPlaying (current State) State {
//         s := current.(*PlayerState)
//         if s.playing {
//                return nil
//         }
//         return &PlayerState{
//                playing: true,
//                song:    s.song,
//        }
//  }
//
// Action is a string, possibly specified by separate
// format+arguments, that identifies what exactly should be done when
// the test generator suggests executing a test step with this
// action. For example, an action can be a keyword or a test step name
// possibly with parameters, or an executable line of any script
// language like "press-button 'play'".
//
// Transition is a combination of an Action and a StateChange
// function. For example,
//
//  play := NewTransition(NewAction("press-button '%s'", "play"), startPlaying)
//
// Finally, Transitions are added to a Model using transition
// generator functions. They are functions that return a slice of
// Transitions that may be specified in a State. For instance:
//
//  model.From(func(current State) []*Transition {
//          return []*Transition{play}
//  })
//
// Note that the transition generator function can already do checks
// on the current State attributes and return only transitions that
// are specified at the state. However, as the StateChange function of
// a returned transition may return nil, not all returned transitions
// need to be defined at the state. Rather, transition generator
// functions enable having common preconditions for all transitions
// that the generator may return. Model.From() can be called multiple
// times to add multiple transition generator functions.
//
// Refer to model_test.go to find examples of defining the same model
// for a player in two different ways: first with StateChanges and
// Transitions, and then with convenience functions When/OnAction/Do.
//
// # Test generation
//
// Tests are sequences of Steps. Every Step has an Action, a start
// State, and an end State. Model.StepsFrom(State) returns all
// possible Steps whose start state is State.
//
// Path is a sequence of Steps where the end state of a Step is the
// start state of the next Step. Model.Paths(State, maxLen) returns
// all possible Paths of at most maxLen Steps where the first Step of
// every Path starts from the State.
//
// Coverer helps finding Paths that increase coverage of wanted
// elements. Elements to be covered are specified by Coverer methods:
//  - CoverStates(): cover unique State.String()s:
//    visit every state.
//  - CoverStateActions(): unique StartState().String() + Action().String():
//    test every action in every state.
//  - CoverStateCombination(n): unique State_1, ..., State_n combinations:
//    test all state-paths of length n.
//  - CoverActions(): unique Action.Strings()s:
//    test every action. Different parameters counts as different actions.
//  - CoverActionCombinations(n): unique Action_1, ..., Action_n combinations:
//    test all action-paths of length n.
//  - CoverActionFormats(): unique Action formats:
//    test every action format, ignoring action parameters.
//
// Calling multiple Cover*() functions allows specifying multiple
// elements whose coverage counts. For example, CoverActions() and
// CoverStates() counts every new Action and every new State, but it
// does not require executing every Action and in every State like
// CoverStateActions() would do. On the other hand, if all these three
// functions are called, then the best test Path is one that gives
// greatest increase the coverage of all the three elements at the
// same time. In practice, this prioritises testing new Actions and
// new States as long as they are found, at the same time when trying
// to cover every Action in every State.
//
// Cover.BestPath(Model, State, maxLen) returns a Path, starting
// from a State in a Model, that results in largest increase in
// whatever elements are covered. The Path is nil if coverage cannot
// be increased by any Path of at most maxLen Steps.
//
// When one or more Steps in a Path have been handled, they are marked
// as covered by calling Coverer.MarkCovered(Step...). Having all
// marked, Coverer.UpdateCoverage() must be called. This can be heavy
// operation, depending on how coverage is measured and how many steps
// have been covered, and therefore it is done separately from marking
// individual Steps covered. Once updated, Coverer.Coverage() returns
// the total number of elements that have been covered by in all
// marked Steps, and Coverer.BestPath() will use new coverage as basis
// when searching for new BestPaths().
//
// Test generation loop example:
//
//  model := myModel()
//  state := &MyModelState { ... } // initial state of generated test
//  coverer := NewCoverer()
//  coverer.CoverActions()
//  coverer.CoverStates()
//  coverer.CoverStateActions()
//  for {
//          path, stats := coverer.BestPath(model, state, 6)
//          if len(path) == 0 {
//                  break // could not find a path that increased coverage
//          }
//          for _, step := range path[:stats.FirstStep+1] {
//                  fmt.Printf("# coverage: %d, step: %s\n", coverer.Coverage(), step)
//                  fmt.Printf("%s\n", step.Action())
//                  coverer.MarkCovered(step)
//                  coverer.UpdateCoverage()
//          }
//          state = path[stats.FirstStep].EndState()
//  }

package gofmbt

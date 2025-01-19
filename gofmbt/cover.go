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
	"math/rand"
	"strings"
)

const (
	BestPathRandomNone = iota
	BestPathRandomAmongEquallyGood
	BestPathRandomAmongFastestMaxCoverageIncrease
	BestPathRandomAmongMaxCoverageIncrease
	BestPathRandomAmongAnyPath
)

// CoveredInPath is a function that returns a slice of strings covered
// by a path.
type CoveredInPath func(Path) []string

// Coverer combines what is counted as covered, how to count it, and
// helps finding Paths that increase coverage.
type Coverer struct {
	coveredPath Path            // Path that is currently covered.
	coverCount  map[string]int  // Strings covered by the coveredPath.
	covFuncs    []CoveredInPath // Functions that return strings covered by a path.
	historyLen  int             // Length of the history in coveredPath that needs to be considered when estimating coverage increase for new steps that extend the path.
	rand        *rand.Rand      // Random number generator initialized with a given seed.
	randomness  int             // Randomness level.
}

// NewCoverer creates a new Coverer.
func NewCoverer() *Coverer {
	return &Coverer{
		coverCount: map[string]int{},
	}
}

// ActionNames returns names of actions in a path.
func ActionNames(path Path) []string {
	names := []string{}
	for _, t := range path {
		names = append(names, t.action.name)
	}
	return names
}

// ActionFormats returns formats of actions in a path.
func ActionFormats(path Path) []string {
	formats := []string{}
	for _, t := range path {
		formats = append(formats, t.action.format)
	}
	return formats
}

// StateStrings returns state names in a path.
func StateStrings(path Path) []string {
	if len(path) == 0 {
		return nil
	}
	s := []string{path[0].start.String()}
	for _, step := range path {
		s = append(s, step.end.String())
	}
	return s
}

// StateActionStrings returns state-action pairs in a path.
func StateActionStrings(path Path) []string {
	stateActionSep := "\x00"
	stateActions := make([]string, 0, len(path))
	for _, step := range path {
		stateActions = append(stateActions, step.StartState().String()+stateActionSep+step.Action().String())
	}
	return stateActions
}

// CoverActions starts counting covered action names.
func (c *Coverer) CoverActions() {
	c.addCovFunc(ActionNames)
}

// CoverActionFormats starts counting covered action formats.
func (c *Coverer) CoverActionFormats() {
	c.addCovFunc(ActionFormats)
}

// CoverActionCombinations starts counting covered action name combinations of length up to combLenMax.
func (c *Coverer) CoverActionCombinations(combLenMax int) {
	if c.historyLen < combLenMax {
		c.historyLen = combLenMax
	}
	actionSep := "\x00"
	c.addCovFunc(func(path Path) []string {
		actionCombs := []string{}
		for combLen := 1; combLen <= combLenMax; combLen++ {
			for first := 0; first <= len(path)-combLen; first++ {
				actionCombs = append(actionCombs, strings.Join(ActionNames(path[first:first+combLen]), actionSep))
			}
		}
		return actionCombs
	})
}

// CoverActionFormatCombinations starts counting covered action format combinations of length up to combLenMax.
func (c *Coverer) CoverActionFormatCombinations(combLenMax int) {
	if c.historyLen < combLenMax {
		c.historyLen = combLenMax
	}
	actionSep := "\x00"
	c.addCovFunc(func(path Path) []string {
		actionCombs := []string{}
		for combLen := 1; combLen <= combLenMax; combLen++ {
			for first := 0; first <= len(path)-combLen; first++ {
				actionCombs = append(actionCombs, strings.Join(ActionFormats(path[first:first+combLen]), actionSep))
			}
		}
		return actionCombs
	})
}

// CoverStates starts counting covered states.
func (c *Coverer) CoverStates() {
	c.addCovFunc(StateStrings)
}

// CoverStateActions starts counting covered state-action pairs.
func (c *Coverer) CoverStateActions() {
	c.addCovFunc(StateActionStrings)
}

// CoverStateCombinations starts counting covered state combinations of length up to combLenMax.
func (c *Coverer) CoverStateCombinations(combLenMax int) {
	if c.historyLen < combLenMax {
		c.historyLen = combLenMax
	}
	stateSep := "\x00"
	c.addCovFunc(func(path Path) []string {
		stateCombs := []string{}
		for combLen := 1; combLen <= combLenMax; combLen++ {
			for first := 0; first <= len(path)-combLen; first++ {
				stateCombs = append(stateCombs, strings.Join(StateStrings(path[first:first+combLen]), stateSep))
			}
		}
		return stateCombs
	})
}

func (c *Coverer) addCovFunc(covFunc CoveredInPath) {
	c.covFuncs = append(c.covFuncs, covFunc)
}

func (c *Coverer) covFunc(path Path) []string {
	allCovered := []string{}
	for _, covFunc := range c.covFuncs {
		allCovered = append(allCovered, covFunc(path)...)
	}
	return allCovered
}

// Coverage returns the number of unique strings covered.
func (c *Coverer) Coverage() int {
	return len(c.coverCount)
}

// CoveredStrings returns all unique strings covered.
func (c *Coverer) CoveredStrings() []string {
	cs := make([]string, 0, len(c.coverCount))
	for s := range c.coverCount {
		cs = append(cs, s)
	}
	return cs
}

// UpdateCoverage updates the count of covered strings.
func (c *Coverer) UpdateCoverage() {
	c.coverCount = map[string]int{}
	for _, s := range c.covFunc(c.coveredPath) {
		c.coverCount[s]++
	}
}

// MarkCovered marks a sequence of steps as covered. The sequence is
// appended to the currently covered path. Note that covered strings
// is not updated until UpdateCoverage() is called.
func (c *Coverer) MarkCovered(step ...*Step) {
	c.coveredPath = append(c.coveredPath, step...)
}

// CoverageIncreaseStats holds statistics on estimated coverage
// increase when extending a path.
type CoverageIncreaseStats struct {
	MaxStep       int // Index of the step in the path extension after which max increase is reached.
	MaxIncrease   int // Maximum increase in coverage with the path extension.
	FirstStep     int // Index of the step in the path extension after which first coverage increase is reached.
	FirstIncrease int // First increase in coverage with the path extension.
}

// EstimateCoverageIncrease estimates coverage increase when extending
// currently coveredPath with a new path.
func (c *Coverer) EstimateCoverageIncrease(path Path) *CoverageIncreaseStats {
	est := &CoverageIncreaseStats{FirstStep: -1, MaxStep: -1}
	fullCoverCount := map[string]int{}
	historyLen := c.historyLen
	if historyLen > len(c.coveredPath) {
		historyLen = len(c.coveredPath)
	}
	pathWithHistory := append(c.coveredPath[len(c.coveredPath)-historyLen:], path...)
	allNewCovered := c.covFunc(pathWithHistory)
	for _, s := range allNewCovered {
		if c.coverCount[s] == 0 {
			fullCoverCount[s]++
		}
	}
	est.MaxIncrease = len(fullCoverCount)
	firstCoverCount := map[string]int{}
	for i := historyLen; i < len(pathWithHistory); i++ {
		newCovered := c.covFunc(pathWithHistory[:i+1])
		for _, s := range newCovered {
			if c.coverCount[s] == 0 {
				firstCoverCount[s]++
			}
		}
		if est.FirstStep == -1 && len(firstCoverCount) > 0 {
			est.FirstStep = i - historyLen
			est.FirstIncrease = len(firstCoverCount)
		}
		if est.MaxStep == -1 && len(firstCoverCount) == est.MaxIncrease {
			est.MaxStep = i - historyLen
			break
		}
	}
	return est
}

// SetBestPathRandom sets the seed and randomness level for selecting
// a path that increase coverage.
func (c *Coverer) SetBestPathRandom(seed int64, randomness int) {
	c.rand = rand.New(rand.NewSource(seed))
	c.randomness = randomness
}

func (c *Coverer) ShufflePaths(paths []Path) {
	if c.rand == nil {
		return
	}
	c.rand.Shuffle(len(paths), func(i, j int) {
		paths[i], paths[j] = paths[j], paths[i]
	})
}

func (c *Coverer) BestPath(m Walkable, s State, maxLen int) (Path, *CoverageIncreaseStats) {
	var best *CoverageIncreaseStats
	var bestPath Path
	paths := m.Paths(s, maxLen)
	if c.randomness != BestPathRandomNone {
		c.ShufflePaths(paths)
	}
	for _, path := range paths {
		est := c.EstimateCoverageIncrease(path)
		// only return paths that increase coverage
		if est.MaxIncrease == 0 {
			continue
		}
		if best == nil {
			bestPath, best = path, est
			if c.randomness >= BestPathRandomAmongAnyPath {
				break
			}
			continue
		}
		if est.MaxIncrease < best.MaxIncrease {
			continue
		}
		if est.MaxIncrease > best.MaxIncrease {
			bestPath, best = path, est
			continue
		}
		// if we are here, est.MaxIncrease == best.MaxIncrease
		if c.randomness == BestPathRandomAmongMaxCoverageIncrease {
			// we are free to take any path with the same max increase, never mind about other stats
			continue
		}
		if est.MaxStep > best.MaxStep {
			continue
		}
		if est.MaxStep < best.MaxStep {
			bestPath, best = path, est
			continue
		}
		// if we are here, est.MaxStep == best.MaxStep
		if c.randomness == BestPathRandomAmongFastestMaxCoverageIncrease {
			// we are free to take any path that equally few steps to reach max increase, never mind about other stats
			continue
		}
		if est.FirstStep > best.FirstStep {
			continue
		}
		if est.FirstStep < best.FirstStep {
			bestPath, best = path, est
			continue
		}
		// if we are here, est.FirstStep == best.FirstStep
		if est.FirstIncrease < best.FirstIncrease {
			continue
		}
		if est.FirstIncrease > best.FirstIncrease {
			bestPath, best = path, est
			continue
		}
		// If we are here, est.FirstIncrease ==
		// best.FirstIncrease that is, est and best paths are
		// equally good. BestPathRandomAmongEquallyGood has
		// been taken care of by shuffling paths in the
		// beginning.
	}
	if len(bestPath) == 0 {
		return nil, nil
	}
	return bestPath, best
}

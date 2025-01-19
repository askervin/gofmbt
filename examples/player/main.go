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

package main

import (
	"fmt"

	m "github.com/askervin/gofmbt/gofmbt"
)

type PlayerState struct {
	playing   bool
	song      int
	songcount int
}

func (ps *PlayerState) String() string {
	desc := "playing"
	if !ps.playing {
		desc = "paused"
	}
	return fmt.Sprintf("%s-song-%v-of-%v", desc, ps.song, ps.songcount)
}

func NewPlayerModel() *m.Model {
	setState := func(playing bool, song, songcount int) m.StateChange {
		return func(_ m.State) m.State {
			return &PlayerState{playing, song, songcount}
		}
	}

	model := m.NewModel()
	model.From(func(start m.State) []*m.Transition {
		s := start.(*PlayerState)
		return m.When(true,
			m.OnAction("reset").Do(setState(false, 1, 1)),
			m.When(s.playing,
				m.OnAction("pause").Do(setState(false, s.song, s.songcount))),
			m.When(!s.playing,
				m.OnAction("play").Do(setState(true, s.song, s.songcount))),
			m.When(s.song < s.songcount,
				m.OnAction("nextsong").Do(setState(s.playing, s.song+1, s.songcount))),
			m.When(s.song > 1,
				m.OnAction("prevsong").Do(setState(s.playing, s.song-1, s.songcount))),
			m.When(s.songcount < 4,
				m.OnAction("addsong(%d)", s.songcount+1).Do(setState(s.playing, s.song, s.songcount+1))),
		)
	})
	return model
}

func main() {
	model := NewPlayerModel()
	coverer := m.NewCoverer()
	// CoverStateActions: from every state test every action
	coverer.CoverStateActions()
	state := &PlayerState{
		playing:   false,
		song:      1,
		songcount: 1,
	}
	stepCount := 0
	for {
		path, stats := coverer.BestPath(model, state, 8)
		if len(path) == 0 {
			fmt.Printf("\n# final coverage: %d, steps: %d\n", coverer.Coverage(), stepCount)
			break
		}
		for _, step := range path[:stats.MaxStep+1] {
			stepCount++
			fmt.Printf("# %d: coverage: %d, state: %s, test: %s\n", stepCount, coverer.Coverage(), step.StartState(), step.Action())
			coverer.MarkCovered(step)
			coverer.UpdateCoverage()
		}
		state = path[stats.FirstStep].EndState().(*PlayerState)
	}
}

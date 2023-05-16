// Copyright 2022 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package worker

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/undistro/zora/pkg/worker/config"
)

func TestDone(t *testing.T) {
	type donepath struct {
		create bool
		dir    bool
		path   string
		mode   os.FileMode
	}

	cases := []struct {
		description string
		donepath    donepath
		done        bool
	}{
		{
			description: "Inexistent 'done' file",
			done:        false,
		},
		{
			description: "File 'done' created",
			donepath: donepath{
				create: true,
				path:   fmt.Sprintf("%s/done", config.DefaultDoneDir),
				mode:   os.FileMode(0644),
			},
			done: true,
		},
		{
			description: "Dir 'done' created",
			donepath: donepath{
				create: true,
				dir:    true,
				path:   fmt.Sprintf("%s/done", config.DefaultDoneDir),
				mode:   os.FileMode(0644),
			},
			done: false,
		},
		{
			description: "File 'done' without read permission",
			donepath: donepath{
				create: true,
				path:   fmt.Sprintf("%s/done", config.DefaultDoneDir),
				mode:   os.FileMode(0000),
			},
			done: true,
		},
	}

	for _, c := range cases {
		if c.donepath.create {
			if !c.donepath.dir {
				if err := os.MkdirAll(path.Dir(c.donepath.path), 0755); err != nil {
					t.Errorf("Setup failed on case: %s\n", c.description)
					t.Fatal(err)
				}
				if err := os.WriteFile(c.donepath.path, []byte{}, c.donepath.mode); err != nil {
					t.Errorf("Setup failed on case: %s\n", c.description)
					t.Fatal(err)
				}
			} else {
				if err := os.MkdirAll(c.donepath.path, 0755); err != nil {
					t.Errorf("Setup failed on case: %s\n", c.description)
					t.Fatal(err)
				}

			}
		}

		if done := Done(c.donepath.path); done != c.done {
			t.Errorf("Case: %s\n", c.description)
			t.Errorf("Expected path <%s> to return <%t>, but got <%t>\n", c.donepath.path, c.done, done)
		}
		if c.donepath.create {
			if err := os.RemoveAll(path.Dir(c.donepath.path)); err != nil {
				t.Errorf("Setup failed on case: %s\n", c.description)
				t.Fatal(err)
			}
		}
	}
}

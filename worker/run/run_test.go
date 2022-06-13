package run

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/getupio-undistro/inspect/worker/config"
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

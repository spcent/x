package uploader

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEngine_TailRun(t *testing.T) {
	// init test data
	tmp := "/tmp/engine-ut/"
	assert.NoError(t, os.RemoveAll(tmp))
	assert.NoError(t, os.Mkdir(tmp, os.FileMode(0755)))
	files := map[string]string{
		"abc1.txt":         "abcabc",
		"abc2.txt":         "112233",
		"abc3.txt":         "445566",
		"exclude/test.txt": "abc123",
	}
	for name, content := range files {
		if strings.HasPrefix(name, "exclude") {
			os.Mkdir(tmp+"exclude", os.FileMode(0744))
		}
		assert.NoError(t, os.WriteFile(tmp+name, []byte(content), os.FileMode(0644)))
	}

	conf := EngineConfig{
		ForceSync: true,
		Excludes:  []string{"exclude"},
	}
	e := NewEngine(conf, &mockUploader{})
	e.TailRun(tmp)
}

func TestEngine_TailRun2(t *testing.T) {
	// init test data
	tmp := "/tmp/engine-ut/"
	assert.NoError(t, os.RemoveAll(tmp))
	assert.NoError(t, os.Mkdir(tmp, os.FileMode(0755)))
	files := map[string]string{
		"abc1.txt":     "abcabc",
		"abc2.txt":     "445566",
		"dir/test.txt": "abc123",
		"failed.txt":   "aaaaaa",
	}
	for name, content := range files {
		if strings.HasPrefix(name, "dir") {
			os.Mkdir(tmp+"dir", os.FileMode(0744))
		}

		assert.NoError(t, os.WriteFile(tmp+name, []byte(content), os.FileMode(0644)))
	}

	e := NewEngine(EngineConfig{}, &mockUploader{})
	e.TailRun(tmp+"abc1.txt", tmp+"abc2.txt", tmp+"dir", tmp+"failed.txt")
}

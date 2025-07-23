package uploader

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockUploader struct {
	listErr   error
	uploadErr error
	deleteErr error
}

func (m *mockUploader) ListObjects(prefix string) ([]Object, error) {
	return []Object{
		{Key: "abc1.txt", ETag: "abc123"},
		{Key: "abc2.txt", ETag: "d0970714757783e6cf17b26fb8e2298f"},
		{Key: "abc4.txt", ETag: "aaa123"},
	}, m.listErr
}

func (m *mockUploader) Upload(object, rawPath string) error {
	if strings.HasSuffix(object, "failed.txt") {
		return fmt.Errorf("test error")
	}

	return m.uploadErr
}

func (m *mockUploader) Delete(object string) error {
	return m.deleteErr
}

func TestSync(t *testing.T) {
	// init test data
	files := map[string]string{
		"abc1.txt": "abcabc",
		"abc2.txt": "112233",
		"abc3.txt": "445566",
	}

	localObjects := make([]Object, 0)
	for name, content := range files {
		hash := md5.Sum([]byte(content))
		localObjects = append(localObjects, Object{
			Key:      name,
			ETag:     hex.EncodeToString(hash[:]),
			FilePath: name,
			Type:     "text/plain",
		})
	}

	// test
	syncer := NewSyncer(&mockUploader{})
	assert.NoError(t, syncer.Sync(localObjects, ""))
}

func TestSync2(t *testing.T) {
	objects := []Object{
		{
			Key:      "test",
			FilePath: "test",
			Type:     "text/plain",
		},
	}

	s := NewSyncer(&mockUploader{listErr: fmt.Errorf("list error")})
	assert.Error(t, s.Sync(objects, ""))

	s = NewSyncer(&mockUploader{uploadErr: fmt.Errorf("upload error")})
	assert.Error(t, s.Sync(objects, ""))

	s = NewSyncer(&mockUploader{deleteErr: fmt.Errorf("delete error")})
	assert.Error(t, s.Sync(objects, ""))
}

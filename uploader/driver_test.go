package uploader

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var driverConfigs = []DriverConfig{
	{
		Name:      "cos",
		Region:    "ap-shanghai",
		Bucket:    "ut-uptoc-1255970412",
		AccessKey: os.Getenv("UPLOADER_COS_AK"),
		SecretKey: os.Getenv("UPLOADER_COS_SK"),
	},
	{
		Name:      "oss",
		Region:    "cn-hangzhou",
		Bucket:    "ut-uptoc",
		AccessKey: os.Getenv("UPLOADER_OSS_AK"),
		SecretKey: os.Getenv("UPLOADER_OSS_SK"),
	},
	{
		Name:      "qiniu",
		Region:    "cn-east-1",
		Bucket:    "ut-uptoc",
		AccessKey: os.Getenv("UPLOADER_QINIU_AK"),
		SecretKey: os.Getenv("UPLOADER_QINIU_SK"),
	},
	{
		Name:      "aws",
		Region:    "ap-northeast-1",
		Bucket:    "ut-uptoc",
		AccessKey: os.Getenv("UPLOADER_S3_AK"),
		SecretKey: os.Getenv("UPLOADER_S3_SK"),
	},
	{
		Name:      "google",
		Region:    "auto",
		Bucket:    "ut-uptoc",
		AccessKey: os.Getenv("UPLOADER_STORAGE_AK"),
		SecretKey: os.Getenv("UPLOADER_STORAGE_SK"),
	},
}

func TestUploader(t *testing.T) {
	tmp := "/tmp/uptoc-driver-ut/"
	assert.NoError(t, os.RemoveAll(tmp))
	assert.NoError(t, os.Mkdir(tmp, os.FileMode(0755)))
	files := map[string]string{
		"abc1.txt": "abcabcabc",
		"abc2.txt": "112233",
		"abc3.txt": "445566",
	}
	for name, content := range files {
		assert.NoError(t, os.WriteFile(tmp+name, []byte(content), os.FileMode(0644)))
	}

	// test the all drivers
	for _, config := range driverConfigs {
		log.Printf("===== driver =====\n%v", config)
		uploader, err := NewDriver(config)
		assert.NoError(t, err)

		// test object upload
		for object := range files {
			assert.NoError(t, uploader.Upload(object, tmp+object))
		}

		// test objects list
		objects, err := uploader.ListObjects("")
		assert.NoError(t, err)
		assert.Equal(t, len(files), len(objects))

		// test object ETag
		for _, object := range objects {
			assert.Equal(t, strings.ToLower(object.ETag), MD5Hex(tmp+object.Key))
		}

		// test object delete
		for object := range files {
			assert.NoError(t, uploader.Delete(object))
		}
	}
}

func TestS3Uploader_Upload(t *testing.T) {
	u, err := NewDriver(driverConfigs[0])
	assert.NoError(t, err)
	assert.Error(t, u.Upload("aaa.txt", "/tmp/abc123/aaa.txt"))
}

func TestNotSupportDriver(t *testing.T) {
	_, err := NewDriver(DriverConfig{
		Name: "abc",
	})
	assert.Error(t, err)
}

func TestDriverValidate(t *testing.T) {
	assert.Error(t, DriverValidate("test"))
	assert.NoError(t, DriverValidate("oss"))
}

package uploader

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// EngineConfig provides core configuration for the engine.
type EngineConfig struct {
	SaveRoot  string   `yaml:"save_root"`
	VisitHost string   `yaml:"visit_host"`
	ForceSync bool     `yaml:"force_sync"`
	Excludes  []string `yaml:"excludes"`
}

// Engine provides the core logic to finish the feature
type Engine struct {
	echo
	conf EngineConfig

	driver Driver
}

// New returns a new engine.
func NewEngine(conf EngineConfig, ud Driver) *Engine {
	return &Engine{
		conf:   conf,
		driver: ud,
	}
}

// TailRun run the core logic with every path.
func (e *Engine) TailRun(paths ...string) {
	for _, path := range paths {
		stat, err := os.Stat(path)
		if err != nil {
			log.Fatalln(err)
		}

		if stat.IsDir() {
			e.uploadDirectory(path)
			continue
		}

		e.uploadFile(path, filepath.Join(e.conf.SaveRoot, stat.Name()))
	}
}

func (e *Engine) uploadDirectory(dirPath string) {
	objects, err := e.loadLocalObjects(dirPath)
	if err != nil {
		log.Fatalln(err)
	}

	// directory sync
	if e.conf.ForceSync {
		s := NewSyncer(e.driver)
		if err := s.Sync(objects, e.conf.SaveRoot); err != nil {
			log.Fatalln(err)
		}
		return
	}

	// directory normal upload
	for _, obj := range objects {
		e.uploadFile(obj.FilePath, obj.Key)
	}
}

func (e *Engine) uploadFile(filePath, object string) {
	if err := e.driver.Upload(object, filePath); err != nil {
		e.Failed(filePath, err)
		return
	}

	e.Success(e.conf.VisitHost, object)
}

func (e *Engine) loadLocalObjects(dirPath string) ([]Object, error) {
	if !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}

	localObjects := make([]Object, 0)
	visitor := func(filePath string, info os.FileInfo, err error) error {
		if os.IsNotExist(err) {
			return err
		}

		if info.IsDir() || e.shouldExclude(dirPath, filePath) {
			return nil
		}

		localPath := strings.TrimPrefix(filePath, dirPath)
		localObjects = append(localObjects, Object{
			Key:      filepath.Join(e.conf.SaveRoot, localPath),
			ETag:     MD5Hex(filePath),
			FilePath: filePath,
		})
		return nil
	}

	if err := filepath.Walk(dirPath, visitor); err != nil {
		return nil, err
	}

	return localObjects, nil
}

func (e *Engine) shouldExclude(dirPath, filePath string) bool {
	parentPath := strings.TrimPrefix(dirPath, "./")
	for _, ePath := range e.conf.Excludes {
		if strings.HasPrefix(filePath, parentPath+strings.TrimPrefix(ePath, "/")) {
			return true
		}
	}

	return false
}

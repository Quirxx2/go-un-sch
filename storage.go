package golangunitedschoolcerts

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/c2fo/vfs/v6"
	"github.com/c2fo/vfs/v6/backend"
	"github.com/c2fo/vfs/v6/backend/mem"
	"github.com/c2fo/vfs/v6/backend/os"
	"github.com/c2fo/vfs/v6/backend/s3"
	"github.com/c2fo/vfs/v6/vfssimple"
)

type Storage interface {
	Add(string, time.Time, *[]byte) error
	Get(string, time.Time) (*[]byte, error)
	Contains(string, time.Time) bool
	Delete(string, time.Time)
	Load() error
}

type VfsStorage struct {
	fs        vfs.FileSystem
	volume    string
	basePath  string
	memCache  Cache[string, certFile]
	diskCache Cache[string, certLink]
}

type certFile struct {
	timestamp time.Time
	file      []byte
}

// Primitive counter for now
func (c certFile) Size() int {
	return 1
}

type certLink struct {
	timestamp time.Time
	// for use with fs directly
	absPath string
	// for use with vfssimple
	uri string
}

// Primitive counter for now
func (c certLink) Size() int {
	return 1
}

func NewVfsStorage(volume string, basePath string, scheme string, opts *vfs.Options) (s *VfsStorage, err error) {
	s = &VfsStorage{
		volume:   volume,
		basePath: basePath,
	}
	if c, err := NewLRUCache[string, certFile](0, nil); err != nil {
		return nil, fmt.Errorf("failed to create memCache: %w", err)
	} else {
		s.memCache = NewSafeCache[string, certFile](c)
	}
	if c, err := NewLRUCache(0, onLinkEviction); err != nil {
		return nil, fmt.Errorf("failed to create diskCache: %w", err)
	} else {
		s.diskCache = NewSafeCache[string, certLink](c)
	}
	if s.fs, err = InitBackend(scheme, opts); err != nil {
		return nil, fmt.Errorf("failed to init backend: %w", err)
	}
	return s, nil
}

func InitBackend(scheme string, opts *vfs.Options) (fs vfs.FileSystem, err error) {
	switch scheme {
	case os.Scheme:
		fs = backend.Backend(os.Scheme)
		if opts != nil {
			return nil, fmt.Errorf("unknown options: %v, for scheme: %v", opts, scheme)
		}
	case s3.Scheme:
		fs = backend.Backend(s3.Scheme)
		if opts != nil {
			fs = fs.(*s3.FileSystem).WithOptions(opts)
		}
	// im memory implementation for testing purposes
	case mem.Scheme:
		fs = backend.Backend(mem.Scheme)
		if opts != nil {
			return nil, fmt.Errorf("unknown options: %v, for scheme: %v", opts, scheme)
		}
	default:
		return nil, fmt.Errorf("unknown scheme: %v", scheme)
	}
	return fs, nil
}

// TODO: add some sort of job queue with retries in another goroutine
func onLinkEviction(key *string, value *certLink) {
	f, _ := vfssimple.NewFile(value.uri)
	f.Delete()
	f.Close()
}

func (s *VfsStorage) Add(id string, timestamp time.Time, cert *[]byte) error {
	cl, ok := s.diskCache.Peek(id)
	if ok && (cl.timestamp.Equal(timestamp) || cl.timestamp.After(timestamp)) {
		return nil
	}
	file, err := s.fs.NewFile(s.volume, s.basePath+id+"_"+timestamp.String()+".pdf")
	if err != nil {
		return fmt.Errorf("failed to initialize file: %w", err)
	}
	_, err = file.Write(*cert)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	s.diskCache.Add(id, certLink{timestamp, file.Path(), file.URI()})
	return nil
}

func (s *VfsStorage) Get(id string, timestamp time.Time) (cert *[]byte, err error) {
	cf, ok := s.memCache.Peek(id)
	if ok && (cf.timestamp.Equal(timestamp) || cf.timestamp.After(timestamp)) {
		s.memCache.Touch(id)
		return &cf.file, nil
	}
	cl, ok := s.diskCache.Peek(id)
	if ok && (cl.timestamp.Equal(timestamp) || cl.timestamp.After(timestamp)) {
		file, err := s.fs.NewFile(s.volume, cl.absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize file: %w", err)
		}
		c, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read from file: %w", err)
		}
		err = file.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close file: %w", err)
		}
		s.diskCache.Touch(id)
		s.memCache.Add(id, certFile{cl.timestamp, c})
		return &c, err
	}
	return nil, fmt.Errorf("no certificate file found for such id: %v and timestamp: %v", id, timestamp)
}

func (s *VfsStorage) Contains(id string, timestamp time.Time) bool {
	cf, ok := s.memCache.Peek(id)
	if ok && (cf.timestamp.Equal(timestamp) || cf.timestamp.After(timestamp)) {
		return true
	}
	cl, ok := s.diskCache.Peek(id)
	if ok && (cl.timestamp.Equal(timestamp) || cl.timestamp.After(timestamp)) {
		return true
	}
	return false
}

func (s *VfsStorage) Delete(id string, timestamp time.Time) {
	cf, ok := s.memCache.Peek(id)
	if ok && (cf.timestamp.Equal(timestamp) || cf.timestamp.Before(timestamp)) {
		s.memCache.Remove(id)
	}
	cl, ok := s.diskCache.Peek(id)
	if ok && (cl.timestamp.Equal(timestamp) || cl.timestamp.Before(timestamp)) {
		s.diskCache.Remove(id)
	}
}

func (s *VfsStorage) Load() error {
	const (
		HEAD = 8
		TAIL = 4
	)
	m := make(map[string][]string)

	loc, err := s.fs.NewLocation(s.volume, s.basePath)
	if err != nil {
		return fmt.Errorf("failed to set up location: %w", err)
	}
	fDir, err := loc.List()
	if err != nil {
		return fmt.Errorf("failed to get file list: %w", err)
	}
	// fill out file list map
	for _, f := range fDir {
		timestamp := f[HEAD+1 : len(f)-TAIL]
		m[timestamp] = append(m[timestamp], f)
	}
	// sort files
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// add files to storage
	layout := "2006-01-02 15:04:05.999999999 -0700 MST"
	for _, timestamp := range keys {
		ts, err := time.Parse(layout, timestamp)
		if err != nil {
			return fmt.Errorf("failed to convert date to time.Time: %w", err)
		}
		for _, mk := range m[timestamp] {
			file, err := loc.NewFile(mk)
			if err != nil {
				return fmt.Errorf("failed to initialize file: %w", err)
			}
			s.diskCache.Add(mk[:HEAD], certLink{ts, file.Path(), file.URI()})
		}
	}
	return nil
}

package golangunitedschoolcerts

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/c2fo/vfs/v6"
	"github.com/c2fo/vfs/v6/backend/mem"
	"github.com/c2fo/vfs/v6/mocks"
	"github.com/c2fo/vfs/v6/vfssimple"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Check that struct implements interface
var _ Storage = &VfsStorage{}

type fsCountingCalls struct {
	vfs.FileSystem
	counter *int
}

func (fs fsCountingCalls) NewFile(volume string, absFilePath string) (vfs.File, error) {
	*fs.counter++
	return fs.FileSystem.NewFile(volume, absFilePath)
}

func createTestStorage(t *testing.T, scheme string, basePath string) (s *VfsStorage) {
	s, err := NewVfsStorage("", basePath, scheme, nil)
	if err != nil {
		assert.FailNow(t, "unexpected error creating inmemory storage: %v", err)
	}
	return s
}

func composeTestCertLink(id string, timestamp time.Time, scheme string, basePath string) *certLink {
	expFileName := "id" + "_" + timestamp.String() + ".pdf"
	return &certLink{
		timestamp: timestamp,
		absPath:   basePath + expFileName,
		uri:       scheme + "://" + basePath + expFileName,
	}
}

func testLinkedCertEqual(t *testing.T, expected []byte, actual *certLink) {
	f, err := vfssimple.NewFile(actual.uri)
	if err != nil {
		assert.FailNow(t, "unexpected error: %v", err)
	}
	b, err := f.Exists()
	if err != nil {
		assert.FailNow(t, "unexpected error: %v", err)
	}
	assert.True(t, b)
	r, err := io.ReadAll(f)
	if err != nil {
		assert.FailNow(t, "unexpected error: %v", err)
	}
	assert.Equal(t, expected, r)
}

func Test_VfsStorage_Add(t *testing.T) {
	scheme := mem.Scheme
	basePath := "/test/"

	t.Run("Add new certificate to the storage", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)

		id := "id"
		expCert := []byte{1, 1, 1, 1}
		now := time.Now()
		expCertLink := composeTestCertLink(id, now, scheme, basePath)

		err := s.Add(id, now, &expCert)
		assert.NoError(t, err)

		v, ok := s.diskCache.Peek("id")
		if assert.True(t, ok) {
			testLinkedCertEqual(t, expCert, v)
		}
		assert.Equal(t, expCertLink, v)
	})

	t.Run("Cache hits after attempt to add same or older certificate", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)
		calls := 0
		s.fs = fsCountingCalls{s.fs, &calls}

		id := "id"
		expCert := []byte{1, 1, 1, 1}
		now := time.Now()
		expCertLink := composeTestCertLink(id, now, scheme, basePath)

		for i := 0; i < 10; i++ {
			err := s.Add(id, now, &expCert)
			assert.NoError(t, err)
		}

		older := now.Add(-1 * time.Hour)
		for i := 0; i < 10; i++ {
			err := s.Add(id, older, &expCert)
			assert.NoError(t, err)
		}

		// check that there was only one fs.NewFile call
		assert.Equal(t, 1, calls)

		v, ok := s.diskCache.Peek(id)
		if assert.True(t, ok) {
			testLinkedCertEqual(t, expCert, v)
		}
		assert.Equal(t, expCertLink, v)
	})

	t.Run("Replace old certificate after attempt to add newer one with same id", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)
		calls := 0
		s.fs = fsCountingCalls{s.fs, &calls}

		id := "id"
		cert := []byte{1, 1, 1, 1}
		now := time.Now()
		oldCertLink := composeTestCertLink(id, now, scheme, basePath)

		err := s.Add(id, now, &cert)
		assert.NoError(t, err)
		// check that there was only one fs.NewFile call
		assert.Equal(t, 1, calls)

		newCert := []byte{2, 2, 2, 2}
		newTime := time.Now()
		expCertLink := composeTestCertLink(id, newTime, scheme, basePath)

		err = s.Add("id", newTime, &newCert)
		assert.NoError(t, err)
		// check that there was only two fs.NewFile calls
		assert.Equal(t, 2, calls)

		v, ok := s.diskCache.Peek("id")
		if assert.True(t, ok) {
			testLinkedCertEqual(t, newCert, expCertLink)
		}
		assert.Equal(t, expCertLink, v)
		assert.Equal(t, 1, s.diskCache.Len())

		// check that old file doesn't exists anymore on vfs backend
		f, err := vfssimple.NewFile(oldCertLink.uri)
		assert.NoError(t, err, "unexpected error")
		b, err := f.Exists()
		assert.NoError(t, err, "unexpected error")
		assert.False(t, b)
	})
	t.Run("vfs return errors", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)

		mockFile := new(mocks.File)
		mockFs := new(mocks.FileSystem)
		mockFs.On("NewFile", mock.Anything, mock.Anything).Return(mockFile, fmt.Errorf("NewFile error"))
		s.fs = mockFs

		err := s.Add("id", time.Now(), &[]byte{})
		assert.ErrorContains(t, err, "NewFile error")

		mockFile = new(mocks.File)
		mockFile.On("Write", mock.Anything).Return(0, fmt.Errorf("Write error"))
		mockFs = new(mocks.FileSystem)
		mockFs.On("NewFile", mock.Anything, mock.Anything).Return(mockFile, nil)
		s.fs = mockFs

		err = s.Add("id", time.Now(), &[]byte{})
		assert.ErrorContains(t, err, "Write error")

		mockFile = new(mocks.File)
		mockFile.On("Write", mock.Anything).Return(0, nil)
		mockFile.On("Close", mock.Anything).Return(fmt.Errorf("Close error"))
		mockFs = new(mocks.FileSystem)
		mockFs.On("NewFile", mock.Anything, mock.Anything).Return(mockFile, nil)
		s.fs = mockFs

		err = s.Add("id", time.Now(), &[]byte{})
		assert.ErrorContains(t, err, "Close error")
	})
}

func Test_VfsStorage_Get(t *testing.T) {
	scheme := mem.Scheme
	basePath := "/test/"
	id := "id"

	t.Run("Get a new certificate from the storage (empty storage)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)
		now := time.Now()

		actCert, err := s.Get(id, now)
		assert.Nil(t, actCert)
		assert.Error(t, err)
	})

	t.Run("Get a new certificate from the storage (memCash used)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)

		expCert := []byte{1, 1, 1, 1}
		now := time.Now()

		s.memCache.Add(id, certFile{now, expCert})
		actCert, err := s.Get(id, now)
		assert.NoError(t, err)
		assert.Equal(t, expCert, *actCert)
	})

	t.Run("Get a new certificate from the storage (diskCash used)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)

		expCert := []byte{1, 1, 1, 1}
		now := time.Now()

		err := s.Add(id, now, &expCert)
		assert.NoError(t, err)
		actCert, err := s.Get(id, now)
		assert.NoError(t, err)
		assert.Equal(t, expCert, *actCert)
	})

	t.Run("File link exists in diskCash but no file in the storage", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)

		now := time.Now()

		s.diskCache.Add(id, certLink{now, "/test/test.pdf", "test.pdf"})
		actCert, err := s.Get(id, now)
		assert.Error(t, err)
		assert.Nil(t, actCert)
	})

	t.Run("Replace an old certificate in the memCash", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)

		oldCf := certFile{
			timestamp: time.Now(),
			file:      []byte{1, 1, 1, 1},
		}
		newCf := certFile{
			timestamp: oldCf.timestamp.Add(1 * time.Hour),
			file:      []byte{2, 2, 2, 2},
		}

		err := s.Add(id, oldCf.timestamp, &oldCf.file)
		assert.NoError(t, err)
		actCert, err := s.Get(id, oldCf.timestamp)
		assert.NoError(t, err)
		assert.Equal(t, oldCf.file, *actCert)
		cl1, ok := s.memCache.Peek(id)
		assert.True(t, ok)
		assert.Equal(t, oldCf, *cl1)

		err = s.Add(id, newCf.timestamp, &newCf.file)
		assert.NoError(t, err)
		actCert, err = s.Get(id, newCf.timestamp)
		assert.NoError(t, err)
		assert.Equal(t, newCf.file, *actCert)
		cl2, ok := s.memCache.Peek(id)
		assert.True(t, ok)
		assert.Equal(t, newCf, *cl2)
	})

	t.Run("vfs return errors", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)
		timestamp := time.Now()

		err := s.Add(id, timestamp, &[]byte{})
		assert.NoError(t, err)

		mockFile := new(mocks.File)
		mockFs := new(mocks.FileSystem)
		mockFs.On("NewFile", mock.Anything, mock.Anything).Return(mockFile, fmt.Errorf("newFile error"))
		s.fs = mockFs

		_, err = s.Get(id, timestamp)
		assert.ErrorContains(t, err, "newFile error")

		mockFile = new(mocks.File)
		mockFile.On("Read", mock.Anything).Return(0, fmt.Errorf("read error"))
		mockFs = new(mocks.FileSystem)
		mockFs.On("NewFile", mock.Anything, mock.Anything).Return(mockFile, nil)
		s.fs = mockFs

		_, err = s.Get(id, timestamp)
		assert.ErrorContains(t, err, "read error")

		mockFile = new(mocks.File)
		mockFile.
			On("Read", mock.Anything).Return(0, io.EOF).
			On("Close").Return(fmt.Errorf("close error"))
		mockFs = new(mocks.FileSystem)
		mockFs.On("NewFile", mock.Anything, mock.Anything).Return(mockFile, nil)
		s.fs = mockFs

		_, err = s.Get(id, timestamp)
		assert.ErrorContains(t, err, "close error")
	})
}

func Test_VfsStorage_Contains(t *testing.T) {
	scheme := mem.Scheme
	basePath := "/test/"
	id := "id"

	t.Run("Check if data with the given id exists (empty storage)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)
		now := time.Now()

		exist := s.Contains(id, now)
		assert.False(t, exist)
	})

	t.Run("Check if data with the given id exists (memCash used)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)
		now := time.Now()

		s.memCache.Add(id, certFile{now, []byte{}})
		exist := s.Contains(id, now)
		assert.True(t, exist)
	})

	t.Run("Check if data with the given id exists (diskCash used / equal dates)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)
		now := time.Now()

		err := s.Add(id, now, &[]byte{})
		assert.NoError(t, err)
		exist := s.Contains(id, now)
		assert.True(t, exist)
	})

	t.Run("Check if data with the given id exists (diskCash used / outdated cert)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)
		now := time.Now()

		err := s.Add(id, now, &[]byte{})
		assert.NoError(t, err)
		exist := s.Contains(id, now.Add(1*time.Hour))
		assert.False(t, exist)
	})

	t.Run("Check if data with the given id exists (diskCash used / actual cert)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)
		now := time.Now()

		err := s.Add(id, now.Add(1*time.Hour), &[]byte{})
		assert.NoError(t, err)
		exist := s.Contains(id, now)
		assert.True(t, exist)
	})
}

func Test_VfsStorage_Delete(t *testing.T) {
	scheme := mem.Scheme
	basePath := "/test/"
	id := "id"

	t.Run("Delete a new certificate from the storage (memCache storage)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)
		now := time.Now()
		expCert := []byte{1, 1, 1, 1}

		s.memCache.Add(id, certFile{now, expCert})
		actCert, err := s.Get(id, now)
		assert.NoError(t, err)
		assert.Equal(t, expCert, *actCert)

		s.Delete(id, now)
		exist := s.Contains(id, now)
		assert.False(t, exist)
	})

	t.Run("Delete a new certificate from the storage (diskCache storage)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)
		now := time.Now()
		expCert := []byte{1, 1, 1, 1}

		err := s.Add(id, now, &expCert)
		assert.NoError(t, err)

		s.Delete(id, now)
		exist := s.Contains(id, now)
		assert.False(t, exist)
	})

	t.Run("Delete a new certificate from both storages", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)
		now := time.Now()
		expCert := []byte{1, 1, 1, 1}

		err := s.Add(id, now, &expCert)
		assert.NoError(t, err)
		actCert, err := s.Get(id, now)
		assert.NoError(t, err)
		assert.Equal(t, expCert, *actCert)

		s.Delete(id, now)
		exist := s.Contains(id, now)
		assert.False(t, exist)
	})

	t.Run("Don't delete a new certificate from both storages with earlier timestamp", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath)
		now := time.Now()
		expCert := []byte{1, 1, 1, 1}

		err := s.Add(id, now, &expCert)
		assert.NoError(t, err)
		actCert, err := s.Get(id, now)
		assert.NoError(t, err)
		assert.Equal(t, expCert, *actCert)

		beforeNow := now.Add(time.Hour * (-1))
		s.Delete(id, beforeNow)
		exist := s.Contains(id, now)
		assert.True(t, exist)
	})
}

func Test_VfsStorage_Load(t *testing.T) {
	scheme := mem.Scheme
	basePath1 := "/test1/"
	basePath2 := "/test2/"
	basePath3 := "/test3/"
	cert := []byte{}
	expNames1 := []string{
		"06e8469f_2022-12-16 15:25:14.059361 +0000 UTC.pdf",
		"10af7531_2022-12-16 15:25:14.079859 +0000 UTC.pdf",
		"617dfc5c_2022-12-16 15:25:14.0299 +0000 UTC.pdf",
	}
	expNames2 := []string{
		"fac0a04c_2022-12-16 15:25:14.057543 +0000 UTC.pdf",
		"06e8469f_2022-12-16 15:25:14.059361 +0000 UTC.pdf",
		"10af7531_2022-12-16 15:25:14.079859 +0000 UTC.pdf",
		"10af7531_2022-12-16 15:25:14.0299 +0000 UTC.pdf",
	}
	expIds1 := []string{
		"06e8469f",
		"10af7531",
		"617dfc5c",
	}
	expIds2 := []string{
		"fac0a04c",
		"06e8469f",
		"10af7531",
	}
	id := "10af7531"
	timestamp := "2022-12-16 15:25:14.079859 +0000 UTC"
	layout := "2006-01-02 15:04:05.999999999 -0700 MST"

	t.Run("Add files from storage fs to diskCache (no errors)", func(t *testing.T) {
		const (
			HEAD = 8
			TAIL = 4
		)
		s := createTestStorage(t, scheme, basePath1)

		for _, n := range expNames1 {
			file, err := s.fs.NewFile(s.volume, s.basePath+n)
			if err != nil {
				assert.FailNow(t, "unexpected error: %v", err)
			}
			_, err = file.Write(cert)
			if err != nil {
				assert.FailNow(t, "unexpected error: %v", err)
			}
			err = file.Close()
			if err != nil {
				assert.FailNow(t, "unexpected error: %v", err)
			}
		}

		err := s.Load()
		assert.NoError(t, err)

		assert.ElementsMatch(t, expIds1, s.diskCache.Keys())

		for _, expName := range expNames1 {
			stamp := expName[HEAD+1 : len(expName)-TAIL]
			ts, err := time.Parse(layout, stamp)
			assert.NoError(t, err)
			ok := s.Contains(id, ts)
			assert.True(t, ok)
		}
	})

	t.Run("Add files from storage fs to diskCache (NewFile error)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath1)

		mockFs := new(mocks.FileSystem)
		mockLoc := new(mocks.Location)
		mockFs.On("NewLocation", mock.Anything, mock.Anything).Return(mockLoc, nil)
		mockLoc.On("List").Return(expNames1, nil)
		mockLoc.On("NewFile", mock.Anything).Return(nil, fmt.Errorf("NewFile error"))
		s.fs = mockFs

		err := s.Load()
		assert.ErrorContains(t, err, "NewFile error")
	})

	t.Run("Add files from storage fs to diskCache (List error)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath1)

		mockFs := new(mocks.FileSystem)
		mockLoc := new(mocks.Location)
		mockFs.On("NewLocation", mock.Anything, mock.Anything).Return(mockLoc, nil)
		mockLoc.On("List").Return([]string{}, fmt.Errorf("List error"))
		s.fs = mockFs

		err := s.Load()
		assert.ErrorContains(t, err, "List error")
	})

	t.Run("Add files from storage fs to diskCache (NewLocation error)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath1)

		mockFs := new(mocks.FileSystem)
		mockLoc := new(mocks.Location)
		mockFs.On("NewLocation", mock.Anything, mock.Anything).Return(mockLoc, fmt.Errorf("NewLocation error"))
		s.fs = mockFs

		err := s.Load()
		assert.ErrorContains(t, err, "NewLocation error")
	})

	t.Run("Add files from storage fs to diskCache (empty storage)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath2)

		err := s.Load()
		assert.NoError(t, err)

		k := s.diskCache.Keys()
		assert.Empty(t, k)
	})

	t.Run("Add files from storage fs to diskCache (equal id's)", func(t *testing.T) {
		s := createTestStorage(t, scheme, basePath3)

		for _, n := range expNames2 {
			file, err := s.fs.NewFile(s.volume, s.basePath+n)
			if err != nil {
				assert.FailNow(t, "unexpected error: %v", err)
			}
			_, err = file.Write(cert)
			if err != nil {
				assert.FailNow(t, "unexpected error: %v", err)
			}
			err = file.Close()
			if err != nil {
				assert.FailNow(t, "unexpected error: %v", err)
			}
		}

		err := s.Load()
		assert.NoError(t, err)

		assert.ElementsMatch(t, expIds2, s.diskCache.Keys())
		ts, err := time.Parse(layout, timestamp)
		assert.NoError(t, err)
		ok := s.Contains(id, ts)
		assert.True(t, ok)
	})
}

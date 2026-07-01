package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type PathKey struct {
	PathName string
	Filename string
}

func (p PathKey) FullPath() string {
	return filepath.Join(p.PathName, p.Filename)
}

type PathTransformFunc func(string) PathKey

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	parts := []string{}
	for i := 0; i < len(hashStr); i += 5 {
		end := i + 5
		if end > len(hashStr) {
			end = len(hashStr)
		}
		parts = append(parts, hashStr[i:end])
	}

	return PathKey{
		PathName: strings.Join(parts, "/"),
		Filename: hashStr,
	}
}

type StoreOpts struct {
	Root              string
	PathTransformFunc PathTransformFunc
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.Root == "" {
		opts.Root = "storage"
	}

	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = CASPathTransformFunc
	}

	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) path(id, key string) string {
	pathKey := s.PathTransformFunc(key)
	return filepath.Join(s.Root, id, pathKey.FullPath())
}

func (s *Store) Has(id, key string) bool {
	_, err := os.Stat(s.path(id, key))
	return err == nil
}

func (s *Store) writeStream(id, key string, r io.Reader) (int64, error) {
	fullPath := s.path(id, key)

	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return 0, err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return io.Copy(f, r)
}

func (s *Store) Write(id, key string, r io.Reader) (int64, error) {
	return s.writeStream(id, key, r)
}

func (s *Store) WriteDecrypt(encKey []byte, id, key string, r io.Reader) (int64, error) {
	fullPath := s.path(id, key)

	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return 0, err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	n, err := copyDecrypt(encKey, r, f)
	return int64(n), err
}

func (s *Store) Read(id, key string) (int64, io.Reader, error) {
	f, err := os.Open(s.path(id, key))
	if err != nil {
		return 0, nil, err
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return 0, nil, err
	}

	return info.Size(), f, nil
}

func (s *Store) Delete(id, key string) error {
	return os.Remove(s.path(id, key))
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}
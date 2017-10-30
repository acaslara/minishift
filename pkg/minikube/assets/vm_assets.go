/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package assets

import (
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
)

type CopyableFile interface {
	io.Reader
	GetLength() int64
	GetAssetName() string
	GetTargetDir() string
	GetTargetName() string
	GetPermissions() string
}

type BaseAsset struct {
	data        []byte
	reader      io.Reader
	Length      int64
	AssetName   string
	TargetDir   string
	TargetName  string
	Permissions string
}

func (b *BaseAsset) GetAssetName() string {
	return b.AssetName
}

func (b *BaseAsset) GetTargetDir() string {
	return b.TargetDir
}

func (b *BaseAsset) GetTargetName() string {
	return b.TargetName
}

func (b *BaseAsset) GetPermissions() string {
	return b.Permissions
}

type FileAsset struct {
	BaseAsset
}

func NewFileAsset(assetName, targetDir, targetName, permissions string) (*FileAsset, error) {
	f := &FileAsset{
		BaseAsset{
			AssetName:   assetName,
			TargetDir:   targetDir,
			TargetName:  targetName,
			Permissions: permissions,
		},
	}
	file, err := os.Open(f.AssetName)
	if err != nil {
		return nil, errors.Wrapf(err, "Error opening file asset: %s", f.AssetName)
	}
	f.reader = file
	return f, nil
}

func (f *FileAsset) GetLength() int64 {
	file, err := os.Open(f.AssetName)
	defer file.Close()
	if err != nil {
		return 0
	}
	fi, err := file.Stat()
	if err != nil {
		return 0
	}
	return int64(fi.Size())
}

func (f *FileAsset) Read(p []byte) (int, error) {
	if f.reader == nil {
		return 0, errors.New("Error attempting FileAsset.Read, FileAsset.reader uninitialized")
	}
	return f.reader.Read(p)
}

type MemoryAsset struct {
	BaseAsset
}

func (m *MemoryAsset) GetLength() int64 {
	return m.Length
}

func (m *MemoryAsset) Read(p []byte) (int, error) {
	return m.reader.Read(p)
}

func CopyFileLocal(f CopyableFile) error {
	if err := os.MkdirAll(f.GetTargetDir(), os.ModePerm); err != nil {
		return errors.Wrapf(err, "error making dirs for %s", f.GetTargetDir())
	}
	targetPath := filepath.Join(f.GetTargetDir(), f.GetTargetName())
	if _, err := os.Stat(targetPath); err == nil {
		if err := os.Remove(targetPath); err != nil {
			return errors.Wrapf(err, "error removing file %s", targetPath)
		}

	}
	target, err := os.Create(targetPath)
	if err != nil {
		return errors.Wrapf(err, "error creating file at %s", targetPath)
	}
	perms, err := strconv.Atoi(f.GetPermissions())
	if err != nil {
		return errors.Wrapf(err, "error converting permissions %s to integer", perms)
	}
	if err := target.Chmod(os.FileMode(perms)); err != nil {
		return errors.Wrapf(err, "error changing file permissions for %s", targetPath)
	}

	if _, err = io.Copy(target, f); err != nil {
		return errors.Wrapf(err, `error copying file %s to target location:
do you have the correct permissions?  The none driver requires sudo for the "start" command`,
			targetPath)
	}
	return target.Close()
}
package util

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Assert is used to enforce a condition is true.
func Assert(cond bool) {
	if !cond {
		panic("assertion failure")
	}
}

// Copy a regular file.
func copyFile(src, dst string) error {
	inputFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	defer inputFile.Close()
	outputFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	return nil
}

// MoveFile moves a file from a source path to destination path.
func MoveFile(ctx context.Context, src string, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	// copy symlink or file
	if info.Mode()&os.ModeSymlink != 0 {
		Debugf(ctx, "Unsupported symlink to %s", dst)
		return nil
	} else if info.Mode().IsRegular() {
		err = copyFile(src, dst)
	} else {
		err = fmt.Errorf("Unsupported file type: `%s`", info.Mode())
	}
	if err != nil {
		return err
	}

	// The copy was successful, so now delete the original file
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("Failed removing original file `%s`: `%s`", src, err)
	}
	return nil
}

// ReadSRISafely reads a cdnjs/sris file safely.
func ReadSRISafely(file string) ([]byte, error) {
	return ReadFileSafely(file, GetSRIsPath())
}

// ReadHumanPackageSafely reads a cdnjs/packages file safely.
func ReadHumanPackageSafely(file string) ([]byte, error) {
	return ReadFileSafely(file, GetHumanPackagesPath())
}

// ReadLibFileSafely reads a cdnjs/cdnjs file safely.
func ReadLibFileSafely(file string) ([]byte, error) {
	return ReadFileSafely(file, GetCDNJSLibrariesPath())
}

// ReadFileSafely reads a cdnjs file from disk safely, checking that
// it is located under the correct directory.
func ReadFileSafely(target, underPath string) ([]byte, error) {
	if !filepath.IsAbs(target) {
		abspath, err := filepath.Abs(target)
		if err != nil {
			return nil, fmt.Errorf("could not get absolute path: %s", err)
		}
		target = abspath
	}

	// check that the target file is located under a particular directory
	if !strings.HasPrefix(target, underPath) {
		return nil, fmt.Errorf("Unsafe file located outside `%s` with path: `%s`", underPath, target)
	}
	return ioutil.ReadFile(target)
}

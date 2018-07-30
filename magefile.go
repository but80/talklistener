// +build mage

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	// "github.com/jessevdk/go-assets"
	"github.com/jessevdk/go-assets"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"golang.org/x/tools/imports"
)

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

var dirStack = []string{}

func pushDir(dir string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}
	dirStack = append(dirStack, wd)
	return nil
}

func popDir() {
	n := len(dirStack)
	if n == 0 {
		return
	}
	_ = os.Chdir(dirStack[n-1])
	dirStack = dirStack[:n-1]
}

// Checkout submodules
func Submodules() error {
	if exists("cmodules/julius/configure") && exists("cmodules/world/makefile") {
		return nil
	}
	if err := sh.RunV("git", "submodule", "init"); err != nil {
		return err
	}
	if err := sh.RunV("git", "submodule", "update"); err != nil {
		return err
	}
	return nil
}

func buildJulius() error {
	if exists("cmodules/julius/libjulius/libjulius.a") {
		return nil
	}
	if err := pushDir("cmodules/julius"); err != nil {
		return err
	}
	defer popDir()
	if err := sh.RunV("./configure"); err != nil {
		return err
	}
	if err := sh.RunV("make"); err != nil {
		return err
	}
	return nil
}

func buildWorld() error {
	if exists("cmodules/world/build/libworld.a") {
		return nil
	}
	if err := pushDir("cmodules/world"); err != nil {
		return err
	}
	defer popDir()
	if err := sh.RunV("make"); err != nil {
		return err
	}
	return nil
}

// Build C modules
func Cmodules() error {
	mg.SerialDeps(Submodules)
	if err := buildJulius(); err != nil {
		return err
	}
	if err := buildWorld(); err != nil {
		return err
	}
	return nil
}

// Generate assets
func Assets() error {
	dst := "internal/assets/assets.go"
	src := []string{
		"cmodules/segmentation-kit/models/hmmdefs_monof_mix16_gid.binhmm",
	}
	if ok, _ := target.Path(dst, src...); ok {
		return nil
	}

	gen := assets.Generator{
		PackageName: "assets",
	}
	for _, s := range src {
		gen.Add(s)
	}
	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := gen.Write(file); err != nil {
		return err
	}
	return nil
}

// Build binary
func Build() error {
	mg.SerialDeps(Cmodules)
	if err := sh.RunV("go", "build", "."); err != nil {
		return err
	}
	return nil
}

// Install binary
func Install() error {
	mg.SerialDeps(Cmodules)
	if err := sh.RunV("go", "install", "."); err != nil {
		return err
	}
	return nil
}

// Run program
func Run() error {
	mg.SerialDeps(Cmodules)
	args := []string{
		"run", "main.go", "generator.go", "f0.go", "fir.go",
		"input/test.wav", "input/test.txt", "output/test.vsqx",
	}
	if err := sh.RunV("go", args...); err != nil {
		return err
	}
	return nil
}

// Run test
func Test() error {
	mg.SerialDeps(Cmodules)
	if err := sh.RunV("go", "test", "./..."); err != nil {
		return err
	}
	return nil
}

// Check coding style
func Lint() error {
	if err := sh.RunV("gometalinter", "--config=.gometalinter.json", "./..."); err != nil {
		return err
	}
	return nil
}

// Format code
func Fmt() error {
	return filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		sep := string(os.PathSeparator)
		if strings.HasPrefix(path, "vendor"+sep) || strings.HasPrefix(path, "cmodules"+sep) {
			return nil
		}
		var in, out []byte
		if in, err = ioutil.ReadFile(path); err != nil {
			return err
		}
		if out, err = imports.Process(path, in, nil); err != nil {
			return err
		}
		if bytes.Equal(in, out) {
			return nil
		}
		return ioutil.WriteFile(path, out, 0644)
	})
}

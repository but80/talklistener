// +build mage

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jessevdk/go-assets"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"golang.org/x/tools/imports"
)

func init() {
	os.Setenv("GOFLAGS", "-mod=vendor")
	os.Setenv("GO111MODULE", "on")
}

var sep = string(filepath.Separator)

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

// サブモジュールのチェックアウト
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

func matchesAll(file string, regex []string) bool {
	code, err := ioutil.ReadFile(file)
	if err != nil {
		return false
	}
	for _, rx := range regex {
		r, err := regexp.Compile(rx)
		if err != nil || !r.Match(code) {
			return false
		}
	}
	return true
}

func replaceAll(file string, fromRegex []string, to []string) error {
	code, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	for i, rx := range fromRegex {
		if len(to) < i+1 {
			continue
		}
		r, err := regexp.Compile(rx)
		if err != nil {
			return err
		}
		code = r.ReplaceAll(code, []byte(to[i]))
	}
	return ioutil.WriteFile(file, code, 0644)
}

// libsent のビルド
func buildSent() error {
	mg.SerialDeps(Submodules)
	dir := filepath.FromSlash("cmodules/julius/libsent")
	src := filepath.Join(dir, "include", "sent", "speech.h")
	dst := filepath.Join(dir, "libsent.a")

	// Juliusが処理可能な語長・音長の制約を拡大
	from := []string{
		`#define\s+MAXSEQNUM\s+\d+`,
		`#define\s+MAXSPEECHLEN\s+\d+`,
	}
	to := []string{
		"#define MAXSEQNUM     1500",
		"#define MAXSPEECHLEN  3200000",
	}

	if exists(dst) && matchesAll(src, to) {
		return nil
	}
	if err := replaceAll(src, from, to); err != nil {
		return err
	}

	// リビルド
	if err := pushDir(dir); err != nil {
		return err
	}
	defer popDir()
	_ = sh.RunV("make", "distclean")
	_ = sh.RunV("." + sep + "configure")
	_ = sh.RunV("make")
	return nil
}

// libjulius のビルド
func buildJulius() error {
	mg.SerialDeps(buildSent)
	if exists("cmodules/julius/libjulius/libjulius.a") {
		return nil
	}
	if err := pushDir("cmodules/julius/libjulius"); err != nil {
		return err
	}
	defer popDir()
	_ = sh.RunV("make", "distclean")
	if err := sh.RunV("./configure"); err != nil {
		return err
	}
	if err := sh.RunV("make"); err != nil {
		return err
	}
	return nil
}

func buildWorld() error {
	mg.SerialDeps(buildJulius)
	if exists("cmodules/world/build/libworld.a") {
		return nil
	}
	if err := pushDir("cmodules/world"); err != nil {
		return err
	}
	defer popDir()
	_ = sh.RunV("make")
	return nil
}

// Cモジュールのビルド
func cmodules() error {
	mg.SerialDeps(buildWorld)
	return nil
}

// Generate assets
func Assets() error {
	dst := "internal/assets/assets.go"
	src := "cmodules/segmentation-kit/models/hmmdefs_monof_mix16_gid.binhmm"
	if ok, _ := target.Path(dst, src); ok {
		return nil
	}

	gen := assets.Generator{
		PackageName: "assets",
	}
	gen.Add(src)
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

// バイナリのビルド
func Build() error {
	mg.SerialDeps(cmodules)
	if err := sh.RunV("go", "build", "."); err != nil {
		return err
	}
	return nil
}

// インストール
func Install() error {
	mg.SerialDeps(cmodules)
	if err := sh.RunV("go", "install", "."); err != nil {
		return err
	}
	return nil
}

// デモの実行
func Run() error {
	// mg.SerialDeps(cmodules)
	args := []string{
		"run", "main.go",
		"-o", "output/test.vsqx",
		"-l", "-v",
		"input/test.wav",
	}
	if err := sh.RunV("go", args...); err != nil {
		return err
	}
	return nil
}

// テストの実行
func Test() error {
	mg.SerialDeps(cmodules)
	if err := sh.RunV("go", "test", "./..."); err != nil {
		return err
	}
	return nil
}

// コーディングスタイルのチェック
func Lint() error {
	if err := sh.RunV("golangci-lint", "run"); err != nil {
		return err
	}
	return nil
}

// コードのフォーマット
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

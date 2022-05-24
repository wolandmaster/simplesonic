package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type errorString struct {
	msg string
}

func NewError(format string, values ...interface{}) error {
	return &errorString{msg: fmt.Sprintf(format, values...)}
}

func (e *errorString) Error() string {
	return UppercaseFirst(e.msg) + "!"
}

func ProcessError(err error) {
	if err != nil {
		log.Panicf("ERROR: %v\n", err)
	}
}

func ProcessErrorArg(arg interface{}, err error) interface{} {
	ProcessError(err)
	return arg
}

func Close(closer io.Closer) {
	ProcessError(closer.Close())
}

func Contains(str string, list ...string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}
	return false
}

func UppercaseFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

func EncodeId(path string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(path))
}

func DecodeId(id string) string {
	return string(ProcessErrorArg(base64.RawURLEncoding.DecodeString(id)).([]byte))
}

func Hash(str string) uint32 {
	hash := fnv.New32a()
	ProcessErrorArg(hash.Write([]byte(str)))
	return hash.Sum32()
}

func ParseNumber(str string) float64 {
	if value, err := strconv.ParseFloat(str, 64); err != nil {
		return math.NaN()
	} else {
		return value
	}
}

func IsExists(path string) bool {
	exists := false
	if _, err := os.Stat(path); err == nil {
		exists = true
	} else if !errors.Is(err, os.ErrNotExist) {
		ProcessError(err)
	}
	return exists
}

func IsAllowedPath(path string) (string, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path, nil
	}
	for _, musicFolder := range Config.MusicFolders {
		if strings.HasPrefix(filepath.Clean(path), filepath.Clean(musicFolder.Path)) {
			return path, nil
		}
	}
	if Config.PlaylistFolder != "" && strings.HasPrefix(filepath.Clean(path), filepath.Clean(Config.PlaylistFolder)) {
		return path, nil
	}
	return "", NewError("access to a path outside the music folders/playlists is prohibited")
}

// DirName returns the same as filepath.Dir without filepath.Clean
func DirName(path string) string {
	vol := filepath.VolumeName(path)
	i := len(path) - 1
	for i >= len(vol) && !os.IsPathSeparator(path[i]) {
		i--
	}
	dir := path[len(vol):i]
	if dir == "" && len(vol) > 2 {
		return vol
	}
	return vol + dir
}

func FileCreateTime(filename string) *DateTime {
	if statx := ProcessErrorArg(Statx(filename)).(*Statx_t); statx != nil && statx.Btime.Sec != 0 {
		return &DateTime{Time: time.Unix(statx.Btime.Sec, int64(statx.Btime.Nsec))}
	}
	return nil
}

func FileChangeTime(filename string) *DateTime {
	return &DateTime{Time: GetFileInfo(filename).ModTime()}
}

type PathInfo struct {
	Parent string
	os.FileInfo
}

func GetFileInfo(path string) os.FileInfo {
	return ProcessErrorArg(os.Stat(path)).(os.FileInfo)
}

func NewPathInfo(parent string, entry os.FileInfo) *PathInfo {
	pathInfo := &PathInfo{Parent: parent, FileInfo: entry}
	if entry.Mode()&os.ModeSymlink != 0 {
		symlinkDest := ProcessErrorArg(os.Readlink(filepath.Join(parent, entry.Name()))).(string)
		if !filepath.IsAbs(symlinkDest) {
			symlinkDest = filepath.Join(parent, symlinkDest)
		}
		musicFolderParts := strings.SplitN(parent, MusicFolderSeparator, 2)
		symlinkDest = strings.Replace(symlinkDest, musicFolderParts[0]+string(os.PathSeparator),
			musicFolderParts[0]+MusicFolderSeparator, 1)
		pathInfo.Parent = DirName(symlinkDest)
		pathInfo.FileInfo = GetFileInfo(symlinkDest)
	}
	return pathInfo
}

func ReadDirSorted(dirname string) []*PathInfo {
	var entries []*PathInfo
	if !IsExists(dirname) {
		return entries
	}
	for _, fileInfo := range ProcessErrorArg(ioutil.ReadDir(dirname)).([]os.FileInfo) {
		entries = append(entries, NewPathInfo(dirname, fileInfo))
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() && !entries[j].IsDir() {
			return true
		} else if !entries[i].IsDir() && entries[j].IsDir() {
			return false
		} else {
			return entries[i].Name() < entries[j].Name()
		}
	})
	return entries
}

func FilterDirEntries(keepFileExt []string, keepFolders bool, entries ...*PathInfo) []*PathInfo {
	var filteredEntries []*PathInfo
	for _, entry := range entries {
		if (entry.IsDir() && keepFolders) ||
			(!entry.IsDir() && len(keepFileExt) == 0) ||
			(!entry.IsDir() && len(keepFileExt) > 0 && Contains(filepath.Ext(entry.Name()), keepFileExt...)) {
			filteredEntries = append(filteredEntries, entry)
		}
	}
	return filteredEntries
}

func Walk(root string, walkFunc func(pathInfo *PathInfo)) {
	walkFunc(NewPathInfo(DirName(root), GetFileInfo(root)))
	var wg sync.WaitGroup
	wg.Add(1)
	queue := make(chan string, 1024)
	queue <- root
	for i := 0; i < runtime.NumCPU()*4; i++ {
		go func() {
			for {
				path, ok := <-queue
				if !ok {
					fmt.Println("stopping worker")
					return
				}
				dir := ProcessErrorArg(os.Open(path)).(*os.File)
				for _, entry := range ProcessErrorArg(dir.Readdir(-1)).([]os.FileInfo) {
					pathInfo := NewPathInfo(path, entry)
					if entry.IsDir() {
						wg.Add(1)
						queue <- pathInfo.Parent + string(os.PathSeparator) + pathInfo.Name()
					}
					walkFunc(pathInfo)
				}
				Close(dir)
				wg.Done()
			}
		}()
	}
	wg.Wait()
	close(queue)
}

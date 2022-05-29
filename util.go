package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
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
	if errors.Is(err, syscall.ECONNRESET) {
		log.Printf("Connection reset by peer: %v\n", err)
	} else if errors.Is(err, syscall.EPIPE) {
		log.Printf("Broken pipe: %v\n", err)
	} else if err != nil {
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

func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
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

func CreateTime(path string) *DateTime {
	if statx := ProcessErrorArg(Statx(path)).(*Statx_t); statx != nil && statx.Btime.Sec != 0 {
		return &DateTime{Time: time.Unix(statx.Btime.Sec, int64(statx.Btime.Nsec))}
	}
	return ChangeTime(path)
}

func ChangeTime(path string) *DateTime {
	return &DateTime{Time: GetFileInfo(path).ModTime()}
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
		symlinkDest = strings.Replace(symlinkDest,
			musicFolderParts[0]+PathSeparator, musicFolderParts[0]+MusicFolderSeparator, 1)
		pathInfo.Parent = DirName(symlinkDest)
		pathInfo.FileInfo = GetFileInfo(symlinkDest)
	}
	return pathInfo
}

type PathInfoList []*PathInfo

func ReadDir(dirname string) *PathInfoList {
	entries := new(PathInfoList)
	if !IsExists(dirname) {
		return entries
	}
	dir := ProcessErrorArg(os.Open(dirname)).(*os.File)
	defer Close(dir)
	for _, entry := range ProcessErrorArg(dir.Readdir(-1)).([]os.FileInfo) {
		*entries = append(*entries, NewPathInfo(dirname, entry))
	}
	return entries
}

func (pathInfoList *PathInfoList) Sort() *PathInfoList {
	sort.Slice(*pathInfoList, func(i, j int) bool {
		if (*pathInfoList)[i].IsDir() && !(*pathInfoList)[j].IsDir() {
			return true
		} else if !(*pathInfoList)[i].IsDir() && (*pathInfoList)[j].IsDir() {
			return false
		} else {
			return (*pathInfoList)[i].Name() < (*pathInfoList)[j].Name()
		}
	})
	return pathInfoList
}

func (pathInfoList *PathInfoList) SortByChild(less func(*Child, *Child) bool) *PathInfoList {
	cache := make(map[*PathInfo]*Child, len(*pathInfoList))
	sort.Slice(*pathInfoList, func(i, j int) bool {
		if _, ok := cache[(*pathInfoList)[i]]; !ok {
			cache[(*pathInfoList)[i]] = BuildChild((*pathInfoList)[i])
		}
		if _, ok := cache[(*pathInfoList)[j]]; !ok {
			cache[(*pathInfoList)[j]] = BuildChild((*pathInfoList)[j])
		}
		return less(cache[(*pathInfoList)[i]], cache[(*pathInfoList)[j]])
	})
	return pathInfoList
}

func (pathInfoList *PathInfoList) Filter(keepFolders bool, keepFileExt ...string) *PathInfoList {
	n := 0
	for _, entry := range *pathInfoList {
		if (entry.IsDir() && keepFolders) || (!entry.IsDir() && Contains(filepath.Ext(entry.Name()), keepFileExt...)) {
			(*pathInfoList)[n] = entry
			n++
		}
	}
	*pathInfoList = (*pathInfoList)[:n]
	return pathInfoList
}

func (pathInfoList *PathInfoList) FilterByChild(filter func(*Child) bool) *PathInfoList {
	n := 0
	for _, entry := range *pathInfoList {
		if filter(BuildChild(entry)) {
			(*pathInfoList)[n] = entry
			n++
		}
	}
	*pathInfoList = (*pathInfoList)[:n]
	return pathInfoList
}

func (pathInfoList *PathInfoList) Shuffle() *PathInfoList {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(*pathInfoList), func(i, j int) {
		(*pathInfoList)[i], (*pathInfoList)[j] = (*pathInfoList)[j], (*pathInfoList)[i]
	})
	return pathInfoList
}

func Walk(root string, walkFunc func(*PathInfo)) {
	walkFunc(NewPathInfo(DirName(root), GetFileInfo(root)))
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	queue := make(chan string, 1024)
	queue <- root
	for i := 0; i < runtime.NumCPU()*4; i++ {
		go func() {
			for {
				if dirname, ok := <-queue; !ok {
					return
				} else {
					for _, entry := range *ReadDir(dirname) {
						if entry.IsDir() {
							waitGroup.Add(1)
							queue <- entry.Parent + PathSeparator + entry.Name()
						}
						walkFunc(entry)
					}
					waitGroup.Done()
				}
			}
		}()
	}
	waitGroup.Wait()
	close(queue)
}

var (
	childCache      = make(map[string]*Child)
	childCacheMutex = sync.RWMutex{}
)

func BuildChild(entry *PathInfo) *Child {
	childPath := entry.Parent + PathSeparator + entry.Name()
	if child, ok := func() (child *Child, ok bool) {
		childCacheMutex.RLock()
		defer childCacheMutex.RUnlock()
		child, ok = childCache[childPath]
		return
	}(); ok && child.Changed.Equal(ChangeTime(childPath).Time) {
		return child
	}
	child := Child{
		Id:      EncodeId(childPath),
		IsDir:   entry.IsDir(),
		Title:   normalizeName(entry.Name()),
		Created: CreateTime(childPath),
		Changed: ChangeTime(childPath),
	}
	childPathParts := getChildPathParts(childPath)
	if len(childPathParts) > 0 && childPathParts[0].IsDir() {
		child.Artist = normalizeName(childPathParts[0].Name())
		if len(childPathParts) == 1 {
			child.Title = child.Artist
		}
	}
	if len(childPathParts) > 1 {
		child.Parent = EncodeId(entry.Parent)
		if childPathParts[1].IsDir() {
			child.Album = normalizeName(childPathParts[1].Name())
			if match := leadingYearRegexp.FindStringSubmatch(child.Album); match != nil {
				child.Album = match[2]
				child.Year = int(ParseNumber(match[1]))
			}
			if len(childPathParts) == 2 {
				child.Title = child.Album
			}
		}
	}
	coverArtFile := childPath + PathSeparator + "folder.jpg"
	if !child.IsDir {
		coverArtFile = entry.Parent + PathSeparator + "folder.jpg"
		child.Suffix = strings.Replace(filepath.Ext(entry.Name()), ".", "", 1)
		child.Size = entry.Size()
		child.Title = child.Title[0 : len(child.Title)-len(filepath.Ext(child.Title))]
		if match := leadingTrackRegexp.FindStringSubmatch(child.Title); match != nil {
			child.Title = match[2]
			child.Track = int(ParseNumber(match[1]))
		}
	}
	if IsExists(coverArtFile) {
		child.CoverArt = EncodeId(coverArtFile)
	}
	childCacheMutex.Lock()
	defer childCacheMutex.Unlock()
	childCache[childPath] = &child
	return &child
}

func getChildPathParts(path string) []os.FileInfo {
	musicFolderParts := strings.SplitN(path, MusicFolderSeparator, 2)
	musicDirectoryParts := strings.Split(musicFolderParts[1], PathSeparator)
	absolutePath := musicFolderParts[0]
	var childPathParts []os.FileInfo
	for _, musicDirectoryPart := range musicDirectoryParts {
		absolutePath = absolutePath + PathSeparator + musicDirectoryPart
		childPathParts = append(childPathParts, GetFileInfo(absolutePath))
	}
	return childPathParts
}

func normalizeName(str string) string {
	return strings.Replace(strings.TrimSpace(str), "_", " ", -1)
}

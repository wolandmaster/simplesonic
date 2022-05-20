package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	trailingYearRegexp  = regexp.MustCompile(`^(.*?)\s*[(\[<{]?\s*((?:19|20)\d{2})\s*[)\]>}]?\s*$`)
	keyValuePairsRegexp = regexp.MustCompile(`(\w+)="([^"]+)"`)
)

type M3U struct{}

type M3UDecoder struct {
	m3uFilename string
	reader      io.Reader
}

type ExtendedPlaylistWithSongs struct {
	Artist        string
	Album         string
	Year          int
	Genre         string
	Images        map[string]string
	MusicBrainzId string
	LastFmId      string
	DiscogsId     string
	SpotifyId     string
	PlaylistWithSongs
}

func ReadPlaylist(filename string) *ExtendedPlaylistWithSongs {
	file := ProcessErrorArg(os.Open(filename)).(*os.File)
	defer Close(file)
	m3u := M3U{}
	playlist := &ExtendedPlaylistWithSongs{}
	ProcessError(m3u.NewDecoder(file).Decode(playlist))
	return playlist
}

func WritePlaylist(filename string, playlist *ExtendedPlaylistWithSongs) {
	file := ProcessErrorArg(os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)).(*os.File)
	defer Close(file)
	m3u := M3U{}
	ProcessErrorArg(file.Write(ProcessErrorArg(m3u.Marshal(playlist)).([]byte)))
	ProcessError(file.Sync())
}

func (*M3U) NewDecoder(file *os.File) *M3UDecoder {
	return &M3UDecoder{
		m3uFilename: file.Name(),
		reader:      file,
	}
}

func (decoder *M3UDecoder) Decode(playlist *ExtendedPlaylistWithSongs) error {
	playlist.Duration = -1
	playlist.Images = make(map[string]string)
	scanner := bufio.NewScanner(decoder.reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#EXTART:") {
			playlist.Artist = strings.TrimSpace(line[8:])
		} else if strings.HasPrefix(line, "#EXTALB:") {
			playlist.Album = strings.TrimSpace(line[8:])
			if match := trailingYearRegexp.FindStringSubmatch(playlist.Album); match != nil {
				playlist.Album = match[1]
				playlist.Year = int(ParseNumber(match[2]))
			}
		} else if strings.HasPrefix(line, "#EXTGENRE:") {
			playlist.Genre = strings.TrimSpace(line[10:])
		} else if strings.HasPrefix(line, "#PLAYLIST:") {
			name, keyValuePairs := removeKeyValuePairs(line[10:])
			playlist.Name = strings.TrimSpace(name)
			playlist.MusicBrainzId = keyValuePairs["musicbrainz"]
			playlist.LastFmId = keyValuePairs["lastfm"]
			playlist.SpotifyId = keyValuePairs["spotify"]
			playlist.DiscogsId = keyValuePairs["discogs"]
		} else if strings.HasPrefix(line, "#EXTIMG:") {
			imageType := strings.TrimSpace(line[8:])
			if scanner.Scan() {
				entry := strings.TrimSpace(scanner.Text())
				if path, err := IsAllowedPath(entry); err == nil && IsExists(path) {
					playlist.Images[imageType] = path
				} else if path, err := IsAllowedPath(DirName(decoder.m3uFilename) +
					string(os.PathSeparator) + entry); err == nil && IsExists(path) {
					playlist.Images[imageType] = path
				}
			}
		} else if strings.HasPrefix(line, "#EXTINF:") {
			trackInfo := strings.SplitN(line[8:], ",", 2)
			if scanner.Scan() {
				entry := strings.TrimSpace(scanner.Text())
				if child := buildPlaylistChild(DirName(decoder.m3uFilename), entry); child != nil {
					child.Title = strings.TrimSpace(trackInfo[1])
					duration, keyValuePairs := removeKeyValuePairs(trackInfo[0])
					child.Duration = int(ParseNumber(strings.TrimSpace(duration)))
					if child.Duration >= 0 {
						if playlist.Duration == -1 {
							playlist.Duration = 0
						}
						playlist.Duration += child.Duration
					}
					if bitrate, ok := keyValuePairs["bitrate"]; ok {
						child.BitRate = int(ParseNumber(bitrate))
					}
					playlist.Entry = append(playlist.Entry, child)
				}
			}
		} else if strings.HasPrefix(line, "#") || line == "" {
			continue
		} else {
			if child := buildPlaylistChild(DirName(decoder.m3uFilename), line); child != nil {
				playlist.Entry = append(playlist.Entry, child)
			}
		}
	}
	playlist.Id = EncodeId(decoder.m3uFilename)
	playlist.SongCount = len(playlist.Entry)
	if playlist.Name == "" {
		playlist.Name = playlist.Album
	}
	if playlist.Name == "" && playlist.SongCount > 0 {
		playlist.Name = playlist.Entry[0].Album
	}
	if createTime := FileCreateTime(decoder.m3uFilename); createTime != nil {
		playlist.Created = createTime
	} else {
		playlist.Created = &DateTime{}
	}
	playlist.Changed = FileChangeTime(decoder.m3uFilename)
	playlist.CoverArt = "pl-" + playlist.Id
	return nil
}

func (*M3U) Marshal(playlist *ExtendedPlaylistWithSongs) ([]byte, error) {
	// TODO
	return []byte{}, nil
}

func (playlist *ExtendedPlaylistWithSongs) GetPlaylistWithSongs() *PlaylistWithSongs {
	for _, entry := range playlist.Entry {
		if playlist.Artist != "" {
			entry.Artist = playlist.Artist
		}
		if playlist.Album != "" {
			entry.Album = playlist.Album
		}
		if playlist.Year > 0 {
			entry.Year = playlist.Year
		}
		if playlist.Genre != "" {
			entry.Genre = playlist.Genre
		}
		if playlist.MusicBrainzId != "" || playlist.LastFmId != "" {
			entry.ArtistId = playlist.Id
			entry.AlbumId = playlist.Id
		}
	}
	return &playlist.PlaylistWithSongs
}

func removeKeyValuePairs(line string) (string, map[string]string) {
	keyValuePairs := make(map[string]string)
	match := keyValuePairsRegexp.FindStringSubmatchIndex(line)
	for match != nil {
		keyValuePairs[strings.ToLower(line[match[2]:match[3]])] = line[match[4]:match[5]]
		line = line[:match[0]] + line[match[1]:]
		match = keyValuePairsRegexp.FindStringSubmatchIndex(line)
	}
	return line, keyValuePairs
}

func buildPlaylistChild(baseDirectory, entry string) *Child {
	var child *Child
	if strings.HasPrefix(entry, "http://") || strings.HasPrefix(entry, "https://") {
		child = &Child{Id: EncodeId(entry)}
	} else {
		if path, err := IsAllowedPath(entry); err == nil && IsExists(path) {
			child = BuildChild(NewPathInfo(DirName(entry), GetFileInfo(path)))
		} else if path, err := IsAllowedPath(filepath.Join(baseDirectory, entry)); err == nil && IsExists(path) {
			child = BuildChild(NewPathInfo(baseDirectory, GetFileInfo(path)))
		}
	}
	return child
}

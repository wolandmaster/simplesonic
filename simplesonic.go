package main

import (
	"fmt"
	"image"
	"log"
	"log/syslog"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	MusicFolderSeparator = string(os.PathSeparator) + "." + string(os.PathSeparator)
)

var (
	musicFileExtensions    = []string{".mp3", ".m4a", ".flac", ".ogg", ".opus", ".oga", ".aac", ".wav", ".wma"}
	videoFileExtensions    = []string{".mp4", ".m4v", ".mpg", ".webm", ".mkv", ".avi", ".wmv", ".flv", ".mov", ".3gp"}
	mediaFileExtensions    = append(musicFileExtensions, videoFileExtensions...)
	playlistFileExtensions = []string{".m3u", ".m3u8"}
	leadingYearRegexp      = regexp.MustCompile(`^[(\[<{]?\s*((?:19|20)\d{2})\s*[)\]>}]?[\s.,-]+(.*)$`)
	leadingTrackRegexp     = regexp.MustCompile(`^(\d{1,3})[\s.,-]+(.*)$`)
)

func main() {
	RegisterHandler("/rest/ping.view", ping)
	RegisterHandler("/rest/getLicense.view", getLicense)
	RegisterHandler("/rest/getMusicFolders.view", getMusicFolders)
	RegisterHandler("/rest/getIndexes.view", getIndexes)
	RegisterHandler("/rest/getMusicDirectory.view", getMusicDirectory)
	RegisterHandler("/rest/getArtistInfo.view", getArtistInfo)
	RegisterHandler("/rest/getRandomSongs.view", getRandomSongs)
	RegisterHandler("/rest/getPlaylists.view", getPlaylists)
	RegisterHandler("/rest/getPlaylist.view", getPlaylist)
	RegisterHandler("/rest/stream.view", stream)
	RegisterHandler("/rest/download.view", stream)
	RegisterHandler("/rest/getCoverArt.view", getCoverArt)
	RegisterHandler("/rest/jukeboxControl.view", jukeboxControl)
	RegisterHandler("/rest/getInternetRadioStations.view", getInternetRadioStations)
	RegisterHandler("/rest/getUser.view", getUser)
	RegisterHandler("/rest/savePlayQueue.view", savePlayQueue)
	RegisterHandler("/rest", unimplemented)
	http.HandleFunc("/", unhandled)
	server := http.Server{
		Addr:         Config.Server.ListenAddress,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if Config.Server.TLSKey != "" && Config.Server.TLSCert != "" {
		log.Panic(server.ListenAndServeTLS(Config.Server.TLSCert, Config.Server.TLSKey))
	} else {
		log.Panic(server.ListenAndServe())
	}
}

func ping(exchange Exchange) {
	exchange.SendResponse()
}

func getLicense(exchange Exchange) {
	exchange.Response.License = &License{Valid: true}
	exchange.SendResponse()
}

func getMusicFolders(exchange Exchange) {
	exchange.Response.MusicFolders = &MusicFolders{}
	for i, musicFolder := range Config.MusicFolders {
		exchange.Response.MusicFolders.MusicFolder = append(
			exchange.Response.MusicFolders.MusicFolder, &MusicFolder{Id: i, Name: musicFolder.Name})
	}
	exchange.SendResponse()
}

func getIndexes(exchange Exchange) {
	exchange.Response.Indexes = &Indexes{LastModified: 0, IgnoredArticles: ""}
	for i, musicFolder := range Config.MusicFolders {
		if Contains(exchange.Request.URL.Query().Get("musicFolderId"), "", strconv.Itoa(i)) {
			for _, entry := range FilterDirEntries(mediaFileExtensions, true,
				ReadDirSorted(filepath.Clean(musicFolder.Path)+string(os.PathSeparator)+".")...) {
				child := BuildChild(entry)
				if child.IsDir {
					exchange.Response.Indexes.AddArtist(&Artist{Id: child.Id, Name: child.Artist})
				} else {
					exchange.Response.Indexes.Child = append(exchange.Response.Indexes.Child, child)
				}
			}
		}
	}
	exchange.Response.Indexes.Sort()
	exchange.SendResponse()
}

func getMusicDirectory(exchange Exchange) {
	if baseDirectory, err := IsAllowedPath(DecodeId(exchange.Request.URL.Query().Get("id"))); err != nil {
		exchange.SendError(0, err.Error())
	} else {
		musicDirectory := BuildChild(NewPathInfo(DirName(baseDirectory), GetFileInfo(baseDirectory)))
		exchange.Response.Directory = &Directory{
			Id:     musicDirectory.Id,
			Parent: musicDirectory.Parent,
			Name:   musicDirectory.Title,
		}
		if playlistFile := baseDirectory + string(os.PathSeparator) + "album.m3u8"; IsExists(playlistFile) {
			playlist := ReadPlaylist(playlistFile).GetPlaylistWithSongs()
			exchange.Response.Directory.Child = playlist.Entry
			exchange.Response.Directory.Name = playlist.Name
		} else {
			for _, entry := range FilterDirEntries(mediaFileExtensions, true, ReadDirSorted(baseDirectory)...) {
				child := BuildChild(entry)
				exchange.Response.Directory.Child = append(exchange.Response.Directory.Child, child)
			}
		}
		if Config.MPD != nil && IsExists(Config.MPD.UnixSocket) {
			for _, child := range exchange.Response.Directory.Child {
				if !child.IsDir && child.Duration == 0 {
					func() {
						mpd := NewMPD(Config.MPD.UnixSocket)
						defer mpd.Disconnect()
						info := mpd.Info(DecodeId(child.Id))
						child.Duration = int(math.Round(ParseNumber(string(info["duration"]))))
					}()
				}
			}
		}
		exchange.SendResponse()
	}
}

func getArtistInfo(exchange Exchange) {
	exchange.Response.ArtistInfo = &ArtistInfo{}
	exchange.SendResponse()
}

func getRandomSongs(exchange Exchange) {
	var songs []*PathInfo
	for i, musicFolder := range Config.MusicFolders {
		if Contains(exchange.Request.URL.Query().Get("musicFolderId"), "", strconv.Itoa(i)) {
			Walk(musicFolder.Path+string(os.PathSeparator)+".", func(pathInfo *PathInfo) {
				if song := FilterDirEntries(musicFileExtensions, false, pathInfo); len(song) == 1 {
					songs = append(songs, song[0])
				}
			})
		}
	}
	size := 10
	if sizeStr := exchange.Request.URL.Query().Get("size"); sizeStr != "" {
		size = int(ParseNumber(sizeStr))
	}
	exchange.Response.RandomSongs = &Songs{}
	for i := 0; i < size; i++ {
		exchange.Response.RandomSongs.Song = append(exchange.Response.RandomSongs.Song,
			BuildChild(songs[rand.Intn(len(songs))]))
	}
	exchange.SendResponse()
}

func getPlaylists(exchange Exchange) {
	exchange.Response.Playlists = &Playlists{}
	userPlaylistFolder := filepath.Clean(Config.PlaylistFolder) +
		MusicFolderSeparator + exchange.Request.URL.Query().Get("u")
	for _, playlistFile := range FilterDirEntries(playlistFileExtensions, false, ReadDirSorted(userPlaylistFolder)...) {
		playlistWithSongs := ReadPlaylist(filepath.Join(userPlaylistFolder, playlistFile.Name())).GetPlaylistWithSongs()
		playlistWithSongs.Owner = exchange.Request.URL.Query().Get("u")
		playlistWithSongs.Public = false
		exchange.Response.Playlists.Playlist = append(exchange.Response.Playlists.Playlist, &playlistWithSongs.Playlist)
	}
	exchange.SendResponse()
}

func getPlaylist(exchange Exchange) {
	if file, err := IsAllowedPath(DecodeId(exchange.Request.URL.Query().Get("id"))); err != nil {
		exchange.SendError(0, err.Error())
	} else {
		exchange.Response.Playlist = ReadPlaylist(file).GetPlaylistWithSongs()
		exchange.SendResponse()
	}
}

func stream(exchange Exchange) {
	if file, err := IsAllowedPath(DecodeId(exchange.Request.URL.Query().Get("id"))); err != nil {
		exchange.SendError(0, err.Error())
	} else {
		exchange.SendFile(file)
	}
}

func getCoverArt(exchange Exchange) {
	coverArtId := exchange.Request.URL.Query().Get("id")
	if strings.HasPrefix(coverArtId, "pl-") {
		coverArtId = coverArtId[3:]
	}
	if file, err := IsAllowedPath(DecodeId(coverArtId)); err != nil {
		exchange.SendError(0, err.Error())
	} else {
		var coverArt image.Image
		if Contains(filepath.Ext(file), playlistFileExtensions...) {
			coverArt = GenerateCover("TODO")
		} else {
			coverArt = OpenImage(file)
		}
		coverArtSize := float64(coverArt.Bounds().Size().X)
		if size := ParseNumber(exchange.Request.URL.Query().Get("size")); !math.IsNaN(size) && size != coverArtSize {
			interpolation := NearestNeighbor
			if size > coverArtSize {
				interpolation = Bilinear
			}
			coverArt = ResizeImage(coverArt, size/coverArtSize, interpolation)
		}
		if Contains(filepath.Ext(file), playlistFileExtensions...) {
			exchange.SendPng(coverArt)
		} else {
			exchange.SendJpeg(coverArt)
		}
	}
}

func jukeboxControl(exchange Exchange) {
	var (
		action = exchange.Request.URL.Query().Get("action")
		index  = exchange.Request.URL.Query().Get("index")
		offset = exchange.Request.URL.Query().Get("offset")
		gain   = exchange.Request.URL.Query().Get("gain")
		files  []string
	)
	for _, id := range exchange.Request.URL.Query()["id"] {
		if file, err := IsAllowedPath(DecodeId(id)); err != nil {
			exchange.SendError(0, err.Error())
			return
		} else {
			files = append(files, file)
		}
	}
	jukebox := NewJukebox()
	defer jukebox.Disconnect()
	switch action {
	case "add":
		jukebox.Add(files...)
	case "set":
		jukebox.Set(files...)
	case "start":
		jukebox.Start()
	case "stop":
		jukebox.Stop()
	case "skip":
		jukebox.Skip(int(ParseNumber(index)), int(ParseNumber(offset)))
	case "clear":
		jukebox.Clear()
	case "remove":
		jukebox.Remove(int(ParseNumber(index)))
	case "shuffle":
		jukebox.Shuffle()
	case "setGain":
		jukebox.SetGain(float32(ParseNumber(gain)))
	}
	if action == "get" {
		exchange.Response.JukeboxPlaylist = jukebox.Playlist()
	} else {
		exchange.Response.JukeboxStatus = jukebox.Status()
	}
	exchange.SendResponse()
}

func getInternetRadioStations(exchange Exchange) {
	exchange.Response.InternetRadioStations = &InternetRadioStations{}
	radioFolder := Config.PlaylistFolder + string(os.PathSeparator) + "_radio"
	for _, radio := range FilterDirEntries(playlistFileExtensions, false, ReadDirSorted(radioFolder)...) {
		playlist := ReadPlaylist(filepath.Join(radioFolder, radio.Name())).GetPlaylistWithSongs()
		for _, entry := range playlist.Entry {
			exchange.Response.InternetRadioStations.InternetRadioStation = append(
				exchange.Response.InternetRadioStations.InternetRadioStation, &InternetRadioStation{
					Id:        entry.Id,
					Name:      entry.Title,
					StreamUrl: DecodeId(entry.Id),
				})
		}
	}
	exchange.SendResponse()
}

func getUser(exchange Exchange) {
	exchange.Response.User = &User{
		Username: exchange.Request.URL.Query().Get("u"), ScrobblingEnabled: false, AdminRole: true,
		SettingsRole: true, DownloadRole: true, UploadRole: true, PlaylistRole: true, CoverArtRole: true,
		CommentRole: true, PodcastRole: true, StreamRole: true, ShareRole: false,
		JukeboxRole: Config.MPD != nil && IsExists(Config.MPD.UnixSocket),
	}
	exchange.SendResponse()
}

func savePlayQueue(exchange Exchange) {
	exchange.SendResponse()
}

func unimplemented(exchange Exchange) {
	exchange.SendError(30, "Not yet implemented!")
}

func unhandled(writer http.ResponseWriter, request *http.Request) {
	message := fmt.Sprintf("Unhandled request: %s from %s", request.URL, request.RemoteAddr)
	log.Printf(message)
	if logger, err := syslog.New(syslog.LOG_ALERT, "simplesonic"); err == nil {
		ProcessError(logger.Alert(message))
	}
	if hijacker, ok := writer.(http.Hijacker); ok {
		if conn, _, err := hijacker.Hijack(); err == nil {
			Close(conn)
		}
	}
}

func BuildChild(entry *PathInfo) *Child {
	childPath := entry.Parent + string(os.PathSeparator) + entry.Name()
	child := Child{
		Id:    EncodeId(childPath),
		IsDir: entry.IsDir(),
		Title: normalizeName(entry.Name()),
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
	coverArtFile := childPath + string(os.PathSeparator) + "folder.jpg"
	if !child.IsDir {
		coverArtFile = entry.Parent + string(os.PathSeparator) + "folder.jpg"
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
	return &child
}

func getChildPathParts(path string) []os.FileInfo {
	musicFolderParts := strings.SplitN(path, MusicFolderSeparator, 2)
	musicDirectoryParts := strings.Split(musicFolderParts[1], string(os.PathSeparator))
	absolutePath := musicFolderParts[0]
	var childPathParts []os.FileInfo
	for _, musicDirectoryPart := range musicDirectoryParts {
		absolutePath = absolutePath + string(os.PathSeparator) + musicDirectoryPart
		childPathParts = append(childPathParts, GetFileInfo(absolutePath))
	}
	return childPathParts
}

func normalizeName(str string) string {
	return strings.Replace(strings.TrimSpace(str), "_", " ", -1)
}

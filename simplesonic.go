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
	PathSeparator        = string(os.PathSeparator)
	MusicFolderSeparator = PathSeparator + "." + PathSeparator
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
	RegisterHandler("/rest/getAlbumList.view", getAlbumList)
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
			for _, entry := range *ReadDir(filepath.Clean(musicFolder.Path)+PathSeparator+".").
				Filter(true, mediaFileExtensions...).Sort() {
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
		if playlistFile := baseDirectory + PathSeparator + "album.m3u8"; IsExists(playlistFile) {
			playlist := ReadPlaylist(playlistFile).GetPlaylistWithSongs()
			exchange.Response.Directory.Child = playlist.Entry
			exchange.Response.Directory.Name = playlist.Name
		} else {
			for _, entry := range *ReadDir(baseDirectory).Filter(true, mediaFileExtensions...).Sort() {
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

func getAlbumList(exchange Exchange) {
	albums := new(PathInfoList)
	for i, musicFolder := range Config.MusicFolders {
		if Contains(exchange.Request.URL.Query().Get("musicFolderId"), "", strconv.Itoa(i)) {
			for _, artist := range *ReadDir(filepath.Clean(musicFolder.Path) + PathSeparator + ".").Filter(true) {
				for _, album := range *ReadDir(artist.Parent + PathSeparator + artist.Name()).Filter(true) {
					*albums = append(*albums, album)
				}
			}
		}
	}
	switch exchange.Request.URL.Query().Get("type") {
	case "alphabeticalByName":
		albums.SortByChild(func(i, j *Child) bool { return i.Album < j.Album })
	case "alphabeticalByArtist":
		albums.SortByChild(func(i, j *Child) bool { return i.Artist < j.Artist })
	case "byYear":
		fromYear := exchange.QueryGetInt("fromYear", 0)
		toYear := exchange.QueryGetInt("toYear", 9999)
		if fromYear < toYear {
			albums.FilterByChild(func(child *Child) bool { return child.Year >= fromYear && child.Year <= toYear }).
				SortByChild(func(i, j *Child) bool { return i.Year < j.Year })
		} else {
			albums.FilterByChild(func(child *Child) bool { return child.Year >= toYear && child.Year <= fromYear }).
				SortByChild(func(i, j *Child) bool { return i.Year > j.Year })
		}
	case "newest":
		albums.SortByChild(func(i, j *Child) bool { return i.Created.After(j.Created.Time) })
	case "random":
		albums.Shuffle()
	case "highest" /* Top rated */, "starred", "recent" /* Recently played */, "frequent" /* Most played */, "byGenre":
		exchange.SendError(30, "Not yet implemented!")
		return
	}
	size := exchange.QueryGetInt("size", 10)
	offset := exchange.QueryGetInt("offset", 0)
	exchange.Response.AlbumList = &AlbumList{}
	for i, offsetEnd := offset, Min(offset+size, len(*albums)-offset); i < offsetEnd; i++ {
		exchange.Response.AlbumList.Album = append(exchange.Response.AlbumList.Album, BuildChild((*albums)[i]))
	}
	exchange.SendResponse()
}

func getRandomSongs(exchange Exchange) {
	var songs PathInfoList
	for i, musicFolder := range Config.MusicFolders {
		if Contains(exchange.Request.URL.Query().Get("musicFolderId"), "", strconv.Itoa(i)) {
			Walk(filepath.Clean(musicFolder.Path)+PathSeparator+".", func(entry *PathInfo) {
				if !entry.IsDir() && Contains(filepath.Ext(entry.Name()), musicFileExtensions...) {
					songs = append(songs, entry)
				}
			})
		}
	}
	exchange.Response.RandomSongs = &Songs{}
	for i, size := 0, exchange.QueryGetInt("size", 10); i < size; i++ {
		exchange.Response.RandomSongs.Song = append(exchange.Response.RandomSongs.Song,
			BuildChild(songs[rand.Intn(len(songs))]))
	}
	exchange.SendResponse()
}

func getPlaylists(exchange Exchange) {
	exchange.Response.Playlists = &Playlists{}
	userPlaylistFolder := filepath.Clean(Config.PlaylistFolder) +
		MusicFolderSeparator + exchange.Request.URL.Query().Get("u")
	for _, playlistFile := range *ReadDir(userPlaylistFolder).Filter(false, playlistFileExtensions...).Sort() {
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
	radioFolder := Config.PlaylistFolder + PathSeparator + "_radio"
	for _, radio := range *ReadDir(radioFolder).Filter(false, playlistFileExtensions...).Sort() {
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

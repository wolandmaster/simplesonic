package main

// http://www.subsonic.org/pages/inc/api/schema/subsonic-rest-api-1.16.1.xsd

import (
	"encoding/xml"
	"sort"
	"time"
	"unicode"
)

const (
	XMLNS                       = "http://subsonic.org/restapi"
	ApiVersion   Version        = "1.16.1"
	OK           ResponseStatus = "ok"
	Failed                      = "failed"
	Music        MediaType      = "music"
	Podcast                     = "podcast"
	Audiobook                   = "audiobook"
	Video                       = "video"
	New          PodcastStatus  = "new"
	Downloading                 = "downloading"
	Completed                   = "completed"
	ErrorPodcast                = "error"
	Deleted                     = "deleted"
	Skipped                     = "skipped"
)

type Response struct {
	XMLName           xml.Name `xml:"subsonic-response" json:"-"`
	XMLNS             string   `xml:"xmlns,attr" json:"-"`
	*SubsonicResponse `json:"subsonic-response"`
}

type SubsonicResponse struct {
	MusicFolders          *MusicFolders          `xml:"musicFolders" json:"musicFolders,omitempty"`
	Indexes               *Indexes               `xml:"indexes" json:"indexes,omitempty"`
	Directory             *Directory             `xml:"directory" json:"directory,omitempty"`
	Genres                *Genres                `xml:"genres" json:"genres,omitempty"`
	Artists               *ArtistsID3            `xml:"artists" json:"artists,omitempty"`
	Artist                *ArtistWithAlbumsID3   `xml:"artist" json:"artist,omitempty"`
	Song                  *Child                 `xml:"song" json:"song,omitempty"`
	Videos                *Videos                `xml:"videos" json:"videos,omitempty"`
	VideoInfo             *VideoInfo             `xml:"videoInfo" json:"videoInfo,omitempty"`
	NowPlaying            *NowPlaying            `xml:"nowPlaying" json:"nowPlaying,omitempty"`
	SearchResult          *SearchResult          `xml:"searchResult" json:"searchResult,omitempty"`
	SearchResult2         *SearchResult2         `xml:"searchResult2" json:"searchResult2,omitempty"`
	SearchResult3         *SearchResult3         `xml:"searchResult3" json:"searchResult3,omitempty"`
	Playlists             *Playlists             `xml:"playlists" json:"playlists,omitempty"`
	Playlist              *PlaylistWithSongs     `xml:"playlist" json:"playlist,omitempty"`
	JukeboxStatus         *JukeboxStatus         `xml:"jukeboxStatus" json:"jukeboxStatus,omitempty"`
	JukeboxPlaylist       *JukeboxPlaylist       `xml:"jukeboxPlaylist" json:"jukeboxPlaylist,omitempty"`
	License               *License               `xml:"license" json:"license,omitempty"`
	Users                 *Users                 `xml:"users" json:"users,omitempty"`
	User                  *User                  `xml:"user" json:"user,omitempty"`
	ChatMessages          *ChatMessages          `xml:"chatMessages" json:"chatMessages,omitempty"`
	AlbumList             *AlbumList             `xml:"albumList" json:"albumList,omitempty"`
	AlbumList2            *AlbumList2            `xml:"albumList2" json:"albumList2,omitempty"`
	RandomSongs           *Songs                 `xml:"randomSongs" json:"randomSongs,omitempty"`
	SongsByGenre          *Songs                 `xml:"songsByGenre" json:"songsByGenre,omitempty"`
	Lyrics                *Lyrics                `xml:"lyrics" json:"lyrics,omitempty"`
	Podcasts              *Podcasts              `xml:"podcasts" json:"podcasts,omitempty"`
	NewestPodcasts        *NewestPodcasts        `xml:"newestPodcasts" json:"newestPodcasts,omitempty"`
	InternetRadioStations *InternetRadioStations `xml:"internetRadioStations" json:"internetRadioStations,omitempty"`
	Bookmarks             *Bookmarks             `xml:"bookmarks" json:"bookmarks,omitempty"`
	PlayQueue             *PlayQueue             `xml:"playQueue" json:"playQueue,omitempty"`
	Shares                *Shares                `xml:"shares" json:"shares,omitempty"`
	Starred               *Starred               `xml:"starred" json:"starred,omitempty"`
	Starred2              *Starred2              `xml:"starred2" json:"starred2,omitempty"`
	AlbumInfo             *AlbumInfo             `xml:"albumInfo" json:"albumInfo,omitempty"`
	ArtistInfo            *ArtistInfo            `xml:"artistInfo" json:"artistInfo,omitempty"`
	ArtistInfo2           *ArtistInfo2           `xml:"artistInfo2" json:"artistInfo2,omitempty"`
	SimilarSongs          *SimilarSongs          `xml:"similarSongs" json:"similarSongs,omitempty"`
	SimilarSongs2         *SimilarSongs2         `xml:"similarSongs2" json:"similarSongs2,omitempty"`
	TopSongs              *TopSongs              `xml:"topSongs" json:"topSongs,omitempty"`
	ScanStatus            *ScanStatus            `xml:"scanStatus" json:"scanStatus,omitempty"`
	Error                 *Error                 `xml:"error" json:"error,omitempty"`
	Status                ResponseStatus         `xml:"status,attr" json:"status"`
	Version               Version                `xml:"version,attr" json:"version"`
}

type ResponseStatus string
type Version string
type UserRating int
type AverageRating float64
type MediaType string
type PodcastStatus string

type DateTime struct {
	time.Time
}

type MusicFolders struct {
	MusicFolder []*MusicFolder `xml:"musicFolder" json:"musicFolder,omitempty"`
}

type MusicFolder struct {
	Id   int    `xml:"id,attr" json:"id"`
	Name string `xml:"name,attr,omitempty" json:"name,omitempty"`
}

type Indexes struct {
	Shortcut        []*Artist `xml:"shortcut" json:"shortcut,omitempty"`
	Index           []*Index  `xml:"index" json:"index,omitempty"`
	Child           []*Child  `xml:"child" json:"child,omitempty"`
	LastModified    int64     `xml:"lastModified,attr" json:"lastModified"`
	IgnoredArticles string    `xml:"ignoredArticles,attr" json:"ignoredArticles"`
}

type Index struct {
	Artist []*Artist `xml:"artist" json:"artist,omitempty"`
	Name   string    `xml:"name,attr" json:"name"`
}

type Artist struct {
	Id             string        `xml:"id,attr" json:"id"`
	Name           string        `xml:"name,attr" json:"name"`
	ArtistImageUrl string        `xml:"artistImageUrl,attr,omitempty" json:"artistImageUrl,omitempty"`
	Starred        *DateTime     `xml:"starred,attr,omitempty" json:"starred,omitempty"`
	UserRating     UserRating    `xml:"userRating,attr,omitempty" json:"userRating,omitempty"`
	AverageRating  AverageRating `xml:"averageRating,attr,omitempty" json:"averageRating,omitempty"`
}

type Genres struct {
	Genre []*Genre `xml:"genre" json:"genre,omitempty"`
}

type Genre struct {
	SongCount  int `xml:"songCount,attr" json:"songCount"`
	AlbumCount int `xml:"albumCount,attr" json:"albumCount"`
}

type ArtistsID3 struct {
	Index           []*IndexID3 `xml:"index" json:"index,omitempty"`
	IgnoredArticles string      `xml:"ignoredArticles,attr" json:"ignoredArticles"`
}

type IndexID3 struct {
	Artist []*ArtistID3 `xml:"artist" json:"artist,omitempty"`
	Name   string       `xml:"name,attr" json:"name"`
}

type ArtistID3 struct {
	Id             string    `xml:"id,attr" json:"id"`
	Name           string    `xml:"name,attr" json:"name"`
	CoverArt       string    `xml:"coverArt,attr,omitempty" json:"coverArt,omitempty"`
	ArtistImageUrl string    `xml:"artistImageUrl,attr,omitempty" json:"artistImageUrl,omitempty"`
	AlbumCount     int       `xml:"albumCount,attr" json:"albumCount"`
	Starred        *DateTime `xml:"starred,attr,omitempty" json:"starred,omitempty"`
}

type ArtistWithAlbumsID3 struct {
	Album []*AlbumID3 `xml:"album" json:"album,omitempty"`
	ArtistID3
}

type AlbumID3 struct {
	Id        string    `xml:"id,attr" json:"id"`
	Name      string    `xml:"name,attr" json:"name"`
	Artist    string    `xml:"artist,attr,omitempty" json:"artist,omitempty"`
	ArtistId  string    `xml:"artistId,attr,omitempty" json:"artistId,omitempty"`
	CoverArt  string    `xml:"coverArt,attr,omitempty" json:"coverArt,omitempty"`
	SongCount int       `xml:"songCount,attr" json:"songCount"`
	Duration  int       `xml:"duration,attr" json:"duration"`
	PlayCount int64     `xml:"playCount,attr,omitempty" json:"playCount,omitempty"`
	Created   *DateTime `xml:"created,attr" json:"created"`
	Starred   *DateTime `xml:"starred,attr,omitempty" json:"starred,omitempty"`
	Year      int       `xml:"year,attr,omitempty" json:"year,omitempty"`
	Genre     string    `xml:"genre,attr,omitempty" json:"genre,omitempty"`
}

type Videos struct {
	Video []*Child `xml:"video" json:"video,omitempty"`
}

type VideoInfo struct {
	Captions   []*Captions        `xml:"captions" json:"captions,omitempty"`
	AudioTrack []*AudioTrack      `xml:"audioTrack" json:"audioTrack,omitempty"`
	Conversion []*VideoConversion `xml:"conversion" json:"conversion,omitempty"`
	Id         string             `xml:"id,attr" json:"id"`
}

type Captions struct {
	Id   string `xml:"id,attr" json:"id"`
	Name string `xml:"name,attr,omitempty" json:"name,omitempty"`
}

type AudioTrack struct {
	Id           string `xml:"id,attr" json:"id"`
	Name         string `xml:"name,attr,omitempty" json:"name,omitempty"`
	LanguageCode string `xml:"languageCode,attr,omitempty" json:"languageCode,omitempty"`
}

type VideoConversion struct {
	Id           string `xml:"id,attr" json:"id"`
	BitRate      int    `xml:"bitRate,attr,omitempty" json:"bitRate,omitempty"`
	AudioTrackId int    `xml:"audioTrackId,attr,omitempty" json:"audioTrackId,omitempty"`
}

type Directory struct {
	Child         []*Child      `xml:"child" json:"child,omitempty"`
	Id            string        `xml:"id,attr" json:"id"`
	Parent        string        `xml:"parent,attr,omitempty" json:"parent,omitempty"`
	Name          string        `xml:"name,attr" json:"name"`
	Starred       *DateTime     `xml:"starred,attr,omitempty" json:"starred,omitempty"`
	UserRating    UserRating    `xml:"userRating,attr,omitempty" json:"userRating,omitempty"`
	AverageRating AverageRating `xml:"averageRating,attr,omitempty" json:"averageRating,omitempty"`
	PlayCount     int64         `xml:"playCount,attr,omitempty" json:"playCount,omitempty"`
}

type Child struct {
	Id                    string        `xml:"id,attr" json:"id"`
	Parent                string        `xml:"parent,attr,omitempty" json:"parent,omitempty"`
	IsDir                 bool          `xml:"isDir,attr" json:"isDir"`
	Title                 string        `xml:"title,attr" json:"title"`
	Album                 string        `xml:"album,attr,omitempty" json:"album,omitempty"`
	Artist                string        `xml:"artist,attr,omitempty" json:"artist,omitempty"`
	Track                 int           `xml:"track,attr,omitempty" json:"track,omitempty"`
	Year                  int           `xml:"year,attr,omitempty" json:"year,omitempty"`
	Genre                 string        `xml:"genre,attr,omitempty" json:"genre,omitempty"`
	CoverArt              string        `xml:"coverArt,attr,omitempty" json:"coverArt,omitempty"`
	Size                  int64         `xml:"size,attr,omitempty" json:"size,omitempty"`
	ContentType           string        `xml:"contentType,attr,omitempty" json:"contentType,omitempty"`
	Suffix                string        `xml:"suffix,attr,omitempty" json:"suffix,omitempty"`
	TranscodedContentType string        `xml:"transcodedContentType,attr,omitempty" json:"transcodedContentType,omitempty"`
	TranscodedSuffix      string        `xml:"transcodedSuffix,attr,omitempty" json:"transcodedSuffix,omitempty"`
	Duration              int           `xml:"duration,attr,omitempty" json:"duration,omitempty"`
	BitRate               int           `xml:"bitRate,attr,omitempty" json:"bitRate,omitempty"`
	Path                  string        `xml:"path,attr,omitempty" json:"path,omitempty"`
	IsVideo               bool          `xml:"isVideo,attr,omitempty" json:"isVideo,omitempty"`
	UserRating            UserRating    `xml:"userRating,attr,omitempty" json:"userRating,omitempty"`
	AverageRating         AverageRating `xml:"averageRating,attr,omitempty" json:"averageRating,omitempty"`
	PlayCount             int64         `xml:"playCount,attr,omitempty" json:"playCount,omitempty"`
	DiscNumber            int           `xml:"discNumber,attr,omitempty" json:"discNumber,omitempty"`
	Created               *DateTime     `xml:"created,attr,omitempty" json:"created,omitempty"`
	Changed               *DateTime     `xml:"-" json:"-"`
	Starred               *DateTime     `xml:"starred,attr,omitempty" json:"starred,omitempty"`
	AlbumId               string        `xml:"albumId,attr,omitempty" json:"albumId,omitempty"`
	ArtistId              string        `xml:"artistId,attr,omitempty" json:"artistId,omitempty"`
	Type                  MediaType     `xml:"type,attr,omitempty" json:"type,omitempty"`
	BookmarkPosition      int64         `xml:"bookmarkPosition,attr,omitempty" json:"bookmarkPosition,omitempty"`
	OriginalWidth         int           `xml:"originalWidth,attr,omitempty" json:"originalWidth,omitempty"`
	OriginalHeight        int           `xml:"originalHeight,attr,omitempty" json:"originalHeight,omitempty"`
}

type NowPlaying struct {
	Entry []*NowPlayingEntry `xml:"entry" json:"entry,omitempty"`
}

type NowPlayingEntry struct {
	Username   string `xml:"username,attr" json:"username"`
	MinutesAgo int    `xml:"minutesAgo,attr" json:"minutesAgo"`
	PlayerId   int    `xml:"playerId,attr" json:"playerId"`
	PlayerName string `xml:"playerName,attr,omitempty" json:"playerName,omitempty"`
	Child
}

type SearchResult struct {
	Match    []*Child `xml:"match" json:"match,omitempty"`
	Offset   int      `xml:"offset,attr" json:"offset"`
	TotalHit int      `xml:"totalHit,attr" json:"totalHit"`
}

type SearchResult2 struct {
	Artist []*Artist `xml:"artist" json:"artist,omitempty"`
	Album  []*Child  `xml:"album" json:"album,omitempty"`
	Song   []*Child  `xml:"song" json:"song,omitempty"`
}

type SearchResult3 struct {
	Artist []*ArtistID3 `xml:"artist" json:"artist,omitempty"`
	Album  []*AlbumID3  `xml:"album" json:"album,omitempty"`
	Song   []*Child     `xml:"song" json:"song,omitempty"`
}

type Playlists struct {
	Playlist []*Playlist `xml:"playlist" json:"playlist,omitempty"`
}

type Playlist struct {
	AllowedUser []string  `xml:"allowedUser" json:"allowedUser,omitempty"`
	Id          string    `xml:"id,attr" json:"id"`
	Name        string    `xml:"name,attr" json:"name"`
	Comment     string    `xml:"comment,attr,omitempty" json:"comment,omitempty"`
	Owner       string    `xml:"owner,attr,omitempty" json:"owner,omitempty"`
	Public      bool      `xml:"public,attr,omitempty" json:"public,omitempty"`
	SongCount   int       `xml:"songCount,attr" json:"songCount"`
	Duration    int       `xml:"duration,attr" json:"duration"`
	Created     *DateTime `xml:"created,attr" json:"created"`
	Changed     *DateTime `xml:"changed,attr" json:"changed"`
	CoverArt    string    `xml:"coverArt,attr,omitempty" json:"coverArt,omitempty"`
}

type PlaylistWithSongs struct {
	Entry []*Child `xml:"entry" json:"entry,omitempty"`
	Playlist
}

type JukeboxStatus struct {
	CurrentIndex int     `xml:"currentIndex,attr" json:"currentIndex"`
	Playing      bool    `xml:"playing,attr" json:"playing"`
	Gain         float32 `xml:"gain,attr" json:"gain"`
	Position     int     `xml:"position,attr,omitempty" json:"position,omitempty"`
}

type JukeboxPlaylist struct {
	Entry []*Child `xml:"entry" json:"entry,omitempty"`
	JukeboxStatus
}

type ChatMessages struct {
	ChatMessage []*ChatMessage `xml:"chatMessage" json:"chatMessage,omitempty"`
}

type ChatMessage struct {
	Username string `xml:"username,attr" json:"username"`
	Time     int64  `xml:"time,attr" json:"time"`
	Message  string `xml:"message,attr" json:"message"`
}

type AlbumList struct {
	Album []*Child `xml:"album" json:"album,omitempty"`
}

type AlbumList2 struct {
	Album []*AlbumID3 `xml:"album" json:"album,omitempty"`
}

type Songs struct {
	Song []*Child `xml:"song" json:"song,omitempty"`
}

type Lyrics struct {
	Artist string `xml:"artist,attr,omitempty" json:"artist,omitempty"`
	Title  string `xml:"title,attr,omitempty" json:"title,omitempty"`
}

type Podcasts struct {
	Channel []*PodcastChannel `xml:"channel" json:"channel,omitempty"`
}

type PodcastChannel struct {
	Episode          []*PodcastEpisode `xml:"episode" json:"episode,omitempty"`
	Id               string            `xml:"id,attr" json:"id"`
	Url              string            `xml:"url,attr" json:"url"`
	Title            string            `xml:"title,attr,omitempty" json:"title,omitempty"`
	Description      string            `xml:"description,attr,omitempty" json:"description,omitempty"`
	CoverArt         string            `xml:"coverArt,attr,omitempty" json:"coverArt,omitempty"`
	OriginalImageUrl string            `xml:"originalImageUrl,attr,omitempty" json:"originalImageUrl,omitempty"`
	Status           PodcastStatus     `xml:"status,attr" json:"status"`
	ErrorMessage     string            `xml:"errorMessage,attr,omitempty" json:"errorMessage,omitempty"`
}

type NewestPodcasts struct {
	Episode []*PodcastEpisode `xml:"episode" json:"episode,omitempty"`
}

type PodcastEpisode struct {
	StreamId    string        `xml:"streamId,attr,omitempty" json:"streamId,omitempty"`
	ChannelId   string        `xml:"channelId,attr" json:"channelId"`
	Description string        `xml:"description,attr,omitempty" json:"description,omitempty"`
	Status      PodcastStatus `xml:"status,attr" json:"status"`
	PublishDate *DateTime     `xml:"publishDate,attr,omitempty" json:"publishDate,omitempty"`
	Child
}

type InternetRadioStations struct {
	InternetRadioStation []*InternetRadioStation `xml:"internetRadioStation" json:"internetRadioStation,omitempty"`
}

type InternetRadioStation struct {
	Id          string `xml:"id,attr" json:"id"`
	Name        string `xml:"name,attr" json:"name"`
	StreamUrl   string `xml:"streamUrl,attr" json:"streamUrl"`
	HomePageUrl string `xml:"homePageUrl,attr,omitempty" json:"homePageUrl,omitempty"`
}

type Bookmarks struct {
	Bookmark []*Bookmark `xml:"bookmark" json:"bookmark,omitempty"`
}

type Bookmark struct {
	Entry    []*Child  `xml:"entry" json:"entry"`
	Position int64     `xml:"position,attr" json:"position"`
	Username string    `xml:"username,attr" json:"username"`
	Comment  string    `xml:"comment,attr,omitempty" json:"comment,omitempty"`
	Created  *DateTime `xml:"created,attr" json:"created"`
	Changed  *DateTime `xml:"changed,attr" json:"changed"`
}

type PlayQueue struct {
	Entry     []*Child  `xml:"entry" json:"entry,omitempty"`
	Current   int       `xml:"current,attr,omitempty" json:"current,omitempty"`
	Position  int64     `xml:"position,attr,omitempty" json:"position,omitempty"`
	Username  string    `xml:"username,attr" json:"username"`
	Changed   *DateTime `xml:"changed,attr" json:"changed"`
	ChangedBy string    `xml:"changedBy,attr" json:"changedBy"`
}

type Shares struct {
	Share []*Share `xml:"share" json:"share,omitempty"`
}

type Share struct {
	Entry       []*Child  `xml:"entry" json:"entry,omitempty"`
	Id          string    `xml:"id,attr" json:"id"`
	Url         string    `xml:"url,attr" json:"url"`
	Description string    `xml:"description,attr,omitempty" json:"description,omitempty"`
	Username    string    `xml:"username,attr" json:"username"`
	Created     *DateTime `xml:"created,attr" json:"created"`
	Expires     *DateTime `xml:"expires,attr,omitempty" json:"expires,omitempty"`
	LastVisited *DateTime `xml:"lastVisited,attr,omitempty" json:"lastVisited,omitempty"`
	VisitCount  int       `xml:"visitCount,attr" json:"visitCount"`
}

type Starred struct {
	Artist []*Artist `xml:"artist" json:"artist,omitempty"`
	Album  []*Child  `xml:"album" json:"album,omitempty"`
	Song   []*Child  `xml:"song" json:"song,omitempty"`
}

type AlbumInfo struct {
	Notes          *string `xml:"notes" json:"notes,omitempty"`
	MusicBrainzId  *string `xml:"musicBrainzId" json:"musicBrainzId,omitempty"`
	LastFmUrl      *string `xml:"lastFmUrl" json:"lastFmUrl,omitempty"`
	SmallImageUrl  *string `xml:"smallImageUrl" json:"smallImageUrl,omitempty"`
	MediumImageUrl *string `xml:"mediumImageUrl" json:"mediumImageUrl,omitempty"`
	LargeImageUrl  *string `xml:"largeImageUrl" json:"largeImageUrl,omitempty"`
}

type ArtistInfoBase struct {
	Biography      *string `xml:"biography" json:"biography,omitempty"`
	MusicBrainzId  *string `xml:"musicBrainzId" json:"musicBrainzId,omitempty"`
	LastFmUrl      *string `xml:"lastFmUrl" json:"lastFmUrl,omitempty"`
	SmallImageUrl  *string `xml:"smallImageUrl" json:"smallImageUrl,omitempty"`
	MediumImageUrl *string `xml:"mediumImageUrl" json:"mediumImageUrl,omitempty"`
	LargeImageUrl  *string `xml:"largeImageUrl" json:"largeImageUrl,omitempty"`
}

type ArtistInfo struct {
	SimilarArtist []*Artist `xml:"similarArtist" json:"similarArtist,omitempty"`
	ArtistInfoBase
}

type ArtistInfo2 struct {
	SimilarArtist []*ArtistID3 `xml:"similarArtist" json:"similarArtist,omitempty"`
	ArtistInfoBase
}

type SimilarSongs struct {
	Song []*Child `xml:"song" json:"song,omitempty"`
}

type SimilarSongs2 struct {
	Song []*Child `xml:"song" json:"song,omitempty"`
}

type TopSongs struct {
	Song []*Child `xml:"song" json:"song,omitempty"`
}

type Starred2 struct {
	Artist []*ArtistID3 `xml:"artist" json:"artist,omitempty"`
	Album  []*AlbumID3  `xml:"album" json:"album,omitempty"`
	Song   []*Child     `xml:"song" json:"song,omitempty"`
}

type License struct {
	Valid          bool      `xml:"valid,attr" json:"valid"`
	Email          string    `xml:"email,attr,omitempty" json:"email,omitempty"`
	LicenseExpires *DateTime `xml:"date,attr,omitempty" json:"date,omitempty"`
	TrialExpires   *DateTime `xml:"trialExpires,attr,omitempty" json:"trialExpires,omitempty"`
}

type ScanStatus struct {
	Scanning bool  `xml:"scanning,attr" json:"scanning"`
	Count    int64 `xml:"count,attr,omitempty" json:"count,omitempty"`
}

type Users struct {
	User []*User `xml:"user" json:"user,omitempty"`
}

type User struct {
	Folder            []*int `xml:"folder" json:"folder,omitempty"`
	Username          string `xml:"username,attr" json:"username"`
	Email             string `xml:"email,attr,omitempty" json:"email,omitempty"`
	ScrobblingEnabled bool   `xml:"scrobblingEnabled,attr" json:"scrobblingEnabled"`
	AdminRole         bool   `xml:"adminRole,attr" json:"adminRole"`
	SettingsRole      bool   `xml:"settingsRole,attr" json:"settingsRole"`
	DownloadRole      bool   `xml:"downloadRole,attr" json:"downloadRole"`
	UploadRole        bool   `xml:"uploadRole,attr" json:"uploadRole"`
	PlaylistRole      bool   `xml:"playlistRole,attr" json:"playlistRole"`
	CoverArtRole      bool   `xml:"coverArtRole,attr" json:"coverArtRole"`
	CommentRole       bool   `xml:"commentRole,attr" json:"commentRole"`
	PodcastRole       bool   `xml:"podcastRole,attr" json:"podcastRole"`
	StreamRole        bool   `xml:"streamRole,attr" json:"streamRole"`
	JukeboxRole       bool   `xml:"jukeboxRole,attr" json:"jukeboxRole"`
	ShareRole         bool   `xml:"shareRole,attr" json:"shareRole"`
}

type Error struct {
	Code    int    `xml:"code,attr" json:"code"`
	Message string `xml:"message,attr,omitempty" json:"message,omitempty"`
}

func NewResponse() *Response {
	return &Response{XMLNS: XMLNS, SubsonicResponse: &SubsonicResponse{Status: OK, Version: ApiVersion}}
}

func (dateTime *DateTime) MarshalText() ([]byte, error) {
	if dateTime == nil {
		return []byte{}, nil
	}
	return []byte(dateTime.Format(time.RFC3339)), nil
}

func (indexes *Indexes) AddArtist(artist *Artist) {
	indexName := artist.IndexName()
	for _, index := range indexes.Index {
		if index.Name == indexName {
			index.Artist = append(index.Artist, artist)
			return
		}
	}
	indexes.Index = append(indexes.Index, &Index{Name: indexName, Artist: []*Artist{artist}})
}

func (indexes *Indexes) Sort() {
	sort.Slice(indexes.Index, func(i, j int) bool {
		return indexes.Index[i].Name < indexes.Index[j].Name
	})
	for _, index := range indexes.Index {
		sort.Slice(index.Artist, func(i, j int) bool {
			return index.Artist[i].Name < index.Artist[j].Name
		})
	}
}

func (artist *Artist) IndexName() string {
	name := rune(artist.Name[0])
	if !unicode.IsLetter(name) {
		return "#"
	}
	return string(unicode.ToUpper(name))
}

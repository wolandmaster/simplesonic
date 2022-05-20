# Simplesonic
- minimalistic Subsonic server [API](http://www.subsonic.org/pages/api.jsp) implementation written in Go
- database free (browsing by folder structure)
- external go package free (using only [standard](https://pkg.go.dev/std@go1.13.15) library)
- jukebox support with [MPD](https://www.musicpd.org/)
- m3u support with [extended](https://en.wikipedia.org/wiki/M3U#Extended_M3U) directives
- tested on [dsub](https://f-droid.org/en/packages/github.daneren2005.dsub/), [subsonic](https://play.google.com/store/apps/details?id=net.sourceforge.subsonic.androidapp)
- for a more feature-rich server, use: [gonic](https://github.com/sentriz/gonic), [airsonic](https://github.com/airsonic-advanced/airsonic-advanced), or [ampache](https://github.com/ampache/ampache) 

## Configuration

### Simplesonic config file _(/etc/simplesonic/simplesonic.json_ or _~/.config/simplesonic/simplesonic.json)_
```
{
  "server": {
    "listenAddress": ":4040"
    "tlsKey": "simplesonic.key"
    "tlsCert": "simplesonic.crt",
  },
  "musicFolders": [
    {
      "name": "Music",
      "path": "/path/to/music"
    },
    {
      "name": "Other",
      "path": "/some/other/folder"
    }
  ],
  "playlistFolder": "/path/to/playlist",
  "users": [
    {
      "username": "alice",
      "password": "********"
    }
  ],  
  "mpd": {
    "unixSocket": "/var/run/mpd.sock"
  }
}
```

### Generate self-signed TLS certificate
```
$ openssl genrsa -out simplesonic.key 2048
$ openssl req -new -x509 -sha256 -key simplesonic.key -out simplesonic.crt -days 3650
```

### Minimal MPD config file _(/etc/mpd.conf_ or _~/.config/mpd/mpd.conf)_
Do **not** set _db_file_ or _music_directory_!
```
bind_to_address    "/var/run/mpd.sock"

input {
        plugin     "curl"       
}

audio_output {                                                                                                                                                                                                                                                                              
        type       "alsa"
        name       "ALSA sound card"
}
```

### Recommended music folder structure
```
music
└── Artist
    └── 1990-Album
        ├── 01-Some_track.mp3
        ├── 02-Some_other_track.mp3
        ├── ...
        ├── album.m3u8                            (optional)
        └── folder.jpg                            (optional)       
```

### Example album.m3u8 file
```
#EXTM3U
#PLAYLIST:Black Ice musicbrainz="701340f6-dea7-3f37-acb6-808950f5299b" lastfm="AC%2FDC/Black+Ice" discogs="8540" spotify="7qVfz4UGONwEd5nQjj0ESN"
#EXTART:AC/DC
#EXTALB:Black Ice (2008)
#EXTGENRE:Rock
#EXTIMG:front cover
folder.jpg
#EXTINF:261 bitrate="320",Rock N Roll Train
01-Rock_N_Roll_Train.mp3
#EXTINF:214 bitrate="320",Skies on Fire
02-Skies_on_Fire.mp3
...
```

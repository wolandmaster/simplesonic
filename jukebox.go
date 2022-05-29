package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/textproto"
	"strings"
)

type Jukebox interface {
	Status() *JukeboxStatus
	Playlist() *JukeboxPlaylist
	Add(files ...string)
	Set(files ...string)
	Start()
	Stop()
	Skip(trackPos int, seconds int)
	Clear()
	Remove(trackPos int)
	Shuffle()
	SetGain(volume float32)
	Disconnect()
}

func NewJukebox() Jukebox {
	return NewMPD(Config.MPD.UnixSocket)
}

// MPD protocol description: https://mpd.readthedocs.io/en/latest/protocol.html
type MPD struct {
	Connection *textproto.Conn
	Version    string
}

type MPDResponse map[string][]byte

func NewMPD(unixSocket string) *MPD {
	connection := ProcessErrorArg(textproto.Dial("unix", unixSocket)).(*textproto.Conn)
	line := ProcessErrorArg(connection.ReadLine()).(string)
	if line[:6] != "OK MPD" {
		log.Panicln("MPD ERROR: No greetings from MPD!")
	}
	return &MPD{Connection: connection, Version: line[7:]}
}

func (mpd *MPD) Status() *JukeboxStatus {
	status := mpd.sendCommand("status")
	jukeboxStatus := &JukeboxStatus{
		CurrentIndex: int(ParseNumber(string(status["song"]))),
		Playing:      string(status["state"]) == "play",
		Gain:         float32(ParseNumber(string(status["volume"])) / 100.0),
		State:        string(status["state"]),
	}
	if elapsed, ok := status["elapsed"]; ok {
		jukeboxStatus.Position = int(math.Round(ParseNumber(string(elapsed))))
	}
	return jukeboxStatus
}

func (mpd *MPD) Playlist() *JukeboxPlaylist {
	jukeboxPlaylist := &JukeboxPlaylist{}
	jukeboxPlaylist.JukeboxStatus = *mpd.Status()
	for _, file := range mpd.getPlaylist() {
		child := BuildChild(NewPathInfo(DirName(file), GetFileInfo(file)))
		jukeboxPlaylist.Entry = append(jukeboxPlaylist.Entry, child)
	}
	return jukeboxPlaylist
}

func (mpd *MPD) Add(files ...string) {
	for _, file := range files {
		mpd.sendCommand(fmt.Sprintf("add %s", mpd.quote(file)))
	}
}

func (mpd *MPD) Set(files ...string) {
	playlist := mpd.getPlaylist()
	for trackPos, file := range files {
		if len(playlist) > trackPos {
			if playlist[trackPos] != file {
				songId := mpd.sendCommand(fmt.Sprintf("addid %s", mpd.quote(file)))["Id"]
				mpd.sendCommand(fmt.Sprintf("delete %d", trackPos))
				mpd.sendCommand(fmt.Sprintf("moveid %s %d", songId, trackPos))
			}
		} else {
			mpd.Add(file)
		}
	}
	for trackPos := len(playlist) - 1; trackPos >= len(files); trackPos-- {
		mpd.Remove(trackPos)
	}
}

func (mpd *MPD) Start() {
	mpd.sendCommand("play")
}

func (mpd *MPD) Stop() {
	mpd.sendCommand("pause 1")
}

func (mpd *MPD) Skip(trackPos, seconds int) {
	mpd.sendCommand(fmt.Sprintf("seek %d %d", trackPos, seconds))
	mpd.Start()
}

func (mpd *MPD) Clear() {
	mpd.sendCommand("clear")
}

func (mpd *MPD) Remove(trackPos int) {
	mpd.sendCommand(fmt.Sprintf("delete %d", trackPos))
}

func (mpd *MPD) Shuffle() {
	mpd.sendCommand("shuffle")
}

func (mpd *MPD) SetGain(volume float32) {
	mpd.sendCommand(fmt.Sprintf("setvol %d", int(volume*100.0)))
}

func (mpd *MPD) Info(file string) MPDResponse {
	return mpd.sendCommand(fmt.Sprintf("lsinfo %s", mpd.quote(file)))
}

func (mpd *MPD) Disconnect() {
	Close(mpd.Connection)
}

func (mpd *MPD) getPlaylist() []string {
	playlistMap := mpd.sendCommand("playlist")
	playlist := make([]string, len(playlistMap))
	for key, entry := range playlistMap {
		playlist[int(ParseNumber(key[:strings.Index(key, ":")]))] = string(entry)
	}
	return playlist
}

func (mpd *MPD) sendCommand(command string) MPDResponse {
	id := ProcessErrorArg(mpd.Connection.Cmd(command)).(uint)
	mpd.Connection.StartResponse(id)
	defer mpd.Connection.EndResponse(id)
	response := make(MPDResponse)
	for {
		line := ProcessErrorArg(mpd.Connection.ReadLine()).(string)
		if line == "OK" {
			break
		} else if strings.HasPrefix(line, "ACK ") {
			log.Panicf("MPD ERROR: %s\n", line[4:])
		} else if strings.HasPrefix(line, "binary: ") {
			data := make([]byte, int(ParseNumber(line[8:])))
			ProcessErrorArg(io.ReadFull(mpd.Connection.R, data))
			ProcessErrorArg(mpd.Connection.R.ReadByte())
			response["binary"] = data
		} else if separatorIndex := strings.Index(line, ": "); separatorIndex > 1 {
			response[line[:separatorIndex]] = []byte(line[separatorIndex+2:])
		}
	}
	return response
}

func (mpd *MPD) quote(str string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "'", "\\'", "\"", "\\\"")
	return "\"" + replacer.Replace(str) + "\""
}

// Description: Tracker request and response handling.
package torrent

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"bit-torrent/bencode"
	"bit-torrent/peers"
)

type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

// buildTrackerURL builds a tracker URL from the torrent file and peer information and returns it as a string.
func (t *TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash": []string{string(t.InfoHash[:])},
		"peer_id":   []string{string(peerID[:])},
		"port":      []string{strconv.Itoa(int(port))},
		"uploaded":  []string{"0"},
		// "downloaded"   : []string{string(t.Length)},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{"0"},
		// "left":       []string{strconv.Itoa(t.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

// requestPeers requests peers from the tracker and returns a slice of peers.
func (t *TorrentFile) requestPeers(peerID [20]byte, port uint16) ([]peers.Peer, error) {
	url, err := t.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}

	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	trackerResp := bencodeTrackerResp{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}

	return peers.Unmarshal([]byte(trackerResp.Peers))
}
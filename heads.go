package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
)

type heads struct {
	primary      *head
	heads        []head
	off          []string
	disconnected []string
}

func newHeads(X *xgb.Conn) heads {
	var primaryHead *head
	var primaryOutput randr.Output

	root := xproto.Setup(X).DefaultScreen(X).Root
	resources, err := randr.GetScreenResourcesCurrent(X, root).Reply()
	if err != nil {
		log.Fatalf("Could not get screen resources: %s.", err)
	}

	primaryOutputReply, _ := randr.GetOutputPrimary(X, root).Reply()
	if primaryOutputReply != nil {
		primaryOutput = primaryOutputReply.Output
	}

	hds := make([]head, 0, len(resources.Outputs))
	off := make([]string, 0)
	disconnected := make([]string, 0)
	for i, output := range resources.Outputs {
		oinfo, err := randr.GetOutputInfo(X, output, 0).Reply()
		if err != nil {
			log.Fatalf("Could not get output info for screen %d: %s.", i, err)
		}
		outputName := string(oinfo.Name)

		if oinfo.Connection != randr.ConnectionConnected {
			disconnected = append(disconnected, outputName)
			continue
		}
		if oinfo.Crtc == 0 {
			off = append(off, outputName)
			continue
		}

		crtcinfo, err := randr.GetCrtcInfo(X, oinfo.Crtc, 0).Reply()
		if err != nil {
			log.Fatalf("Could not get crtc info for screen (%d, %s): %s.",
				i, outputName, err)
		}

		head := newHead(output, outputName, crtcinfo)
		if output == primaryOutput {
			primaryHead = &head
		}
		hds = append(hds, head)
	}
	if primaryHead == nil && len(hds) > 0 {
		tmp := hds[0]
		primaryHead = &tmp
	}

	hdsPrim := heads{
		primary:      primaryHead,
		heads:        hds,
		off:          off,
		disconnected: disconnected,
	}
	sort.Sort(hdsPrim)
	return hdsPrim
}

func (hs heads) connectedRandrName(config *config, queryName string) string {
	if hd := hs.findActive(config, queryName); hd != nil {
		return hd.output
	}
	if hdName := hs.findOff(config, queryName); len(hdName) > 0 {
		return hdName
	}
	return ""
}

func (hs heads) findActive(config *config, queryName string) *head {
	if queryName == "primary" {
		return hs.primary
	}
	for _, hd := range hs.heads {
		if hd.output == queryName || config.nice(hd.output) == queryName {
			return &hd
		}
	}
	return nil
}

func (hs heads) findOff(config *config, queryName string) string {
	for _, output := range hs.off {
		if output == queryName || config.nice(output) == queryName {
			return output
		}
	}
	return ""
}

func (hs heads) findDisconnected(config *config, queryName string) string {
	for _, output := range hs.disconnected {
		if output == queryName || config.nice(output) == queryName {
			return output
		}
	}
	return ""
}

func (hs heads) String() string {
	lines := make([]string, len(hs.heads))
	for i, head := range hs.heads {
		lines[i] = fmt.Sprintf("%d: %s (%d, %d) %dx%d",
			i, head.output, head.x, head.y, head.width, head.height)
	}
	return strings.Join(lines, "\n")
}

func (hs heads) Len() int {
	return len(hs.heads)
}

func (hs heads) Less(i, j int) bool {
	h1, h2 := hs.heads[i], hs.heads[j]
	return h1.x < h2.x || (h1.x == h2.x && h1.y < h2.y)
}

func (hs heads) Swap(i, j int) {
	hs.heads[i], hs.heads[j] = hs.heads[j], hs.heads[i]
}

type head struct {
	id                  randr.Output
	output              string
	x, y, width, height int
}

func newHead(id randr.Output, name string, info *randr.GetCrtcInfoReply) head {
	return head{
		id:     id,
		output: name,
		x:      int(info.X),
		y:      int(info.Y),
		width:  int(info.Width),
		height: int(info.Height),
	}
}

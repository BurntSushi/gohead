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
	primary *head
	heads   []head
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
	for i, output := range resources.Outputs {
		oinfo, err := randr.GetOutputInfo(X, output, 0).Reply()
		if err != nil {
			log.Fatalf("Could not get output info for screen %d: %s.", i, err)
		}
		if oinfo.Connection != randr.ConnectionConnected {
			continue
		}

		outputName := string(oinfo.Name)
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
	if primaryHead == nil {
		tmp := hds[0]
		primaryHead = &tmp
	}

	hdsPrim := heads{
		primary: primaryHead,
		heads:   hds,
	}
	sort.Sort(hdsPrim)
	return hdsPrim
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

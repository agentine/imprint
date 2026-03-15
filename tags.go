package imprint

import "strings"

type tagOpts struct {
	ignore    bool
	set       bool
	str       bool
	omitempty bool
}

func parseTag(tag string) tagOpts {
	var t tagOpts
	for _, part := range strings.Split(tag, ",") {
		switch strings.TrimSpace(part) {
		case "-", "ignore":
			t.ignore = true
		case "set":
			t.set = true
		case "string":
			t.str = true
		case "omitempty":
			t.omitempty = true
		}
	}
	return t
}

package netconf

import (
	"io/ioutil"
	"testing"

	"github.com/m-lab/go/rtx"
)

func TestCompare(t *testing.T) {
	abc01Conf, err := ioutil.ReadFile("testdata/abc01.conf")
	rtx.Must(err, "Cannot read test data")
	abc01BisConf, err := ioutil.ReadFile("testdata/abc01-bis.conf")
	rtx.Must(err, "Cannot read test data")
	abc02Conf, err := ioutil.ReadFile("testdata/abc02.conf")
	rtx.Must(err, "Cannot read test data")

	tests := []struct {
		name string
		c1   string
		c2   string
		want bool
	}{
		{
			name: "empty-string",
			c1:   "",
			c2:   "",
			want: true,
		},
		{
			name: "comments-removed",
			c1:   "# this is a comment",
			c2:   "# this is another comment",
			want: true,
		},
		{
			name: "single-line-comments",
			c1:   "#comment\n#another comment",
			c2:   "#more comments\n#",
			want: true,
		},
		{
			name: "single-line-comments-2",
			c1:   "# comment\nline\n#comment\nline",
			c2:   "line\nline",
			want: true,
		},
		{
			name: "c-style-comments",
			c1:   "text/* comment */text\n",
			c2:   "texttext",
			want: true,
		},
		{
			name: "c-style-multi-line-comments",
			c1:   "text\n/*multi\nline\ncomment*/text",
			c2:   "text\ntext",
			want: true,
		},
		{
			name: "version-removed",
			c1:   "version 1",
			c2:   "version 2",
			want: true,
		},
		{
			name: "multiline-version",
			c1:   "version 1\nversion 2",
			c2:   "version 2\nversion 1",
			want: true,
		},
		{
			name: "comments-and-version",
			c1:   "#comment\nversion 1",
			c2:   "version 1\n##comment",
			want: true,
		},
		{
			name: "different-configs",
			c1:   string(abc01Conf),
			c2:   string(abc02Conf),
			want: false,
		},
		{
			name: "same-configs-different-comments",
			c1:   string(abc01Conf),
			c2:   string(abc01BisConf),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Compare(tt.c1, tt.c2); got != tt.want {
				t.Errorf("Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

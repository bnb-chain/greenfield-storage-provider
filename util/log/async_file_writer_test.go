package log

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestWriter(t *testing.T) {
	w := NewAsyncFileWriter("./hello.log", 100)
	w.Write([]byte("hello\n"))
	w.Write([]byte("world\n"))
	w.Stop()
	files, _ := ioutil.ReadDir("./")
	for _, f := range files {
		fn := f.Name()
		if strings.HasPrefix(fn, "hello") {
			t.Log(fn)
			content, _ := ioutil.ReadFile(fn)
			t.Log(string(content))
			os.Remove(fn)
		}
	}
}

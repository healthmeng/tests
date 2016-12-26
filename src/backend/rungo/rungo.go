package rungo

import (
	"backend/runsrc"
	"fmt"
	"os"
	"strings"
)

func init() {
	runsrc.Register("GoLang", Judgement, RetCmd)
}

func Judgement(srcpath string) int {
	if strings.HasSuffix(srcpath, ".go") ||
		strings.HasSuffix(srcpath, ".GO") {
		return 1
	} else {
		return 0
	}
}

func RetCmd(tp int, srcpath string, args ...string) string {
	st, _ := os.Stat(srcpath)
	ret := fmt.Sprintf("/usr/bin/go run /tmp/%s", st.Name())
	for _, arg := range args {
		ret += " " + arg
	}
	return ret
}

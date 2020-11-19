package runasm

import (
	"backend/runsrc"
	"fmt"
	"os"
	"strings"
)

func init() {
	runsrc.Register("ASM", Judgement, RetCmd)
}

func Judgement(srcpath string) int {
	lowername:=strings.ToLower(srcpath)
	if strings.HasSuffix(lowername, ".asm") || strings.HasSuffix(lowername, ".s") {
		return 1
	} else {
		return 0
	}
}

func RetCmd(tp int, srcpath string, args ...string) string {
	st, _ := os.Stat(srcpath)
	compiler := "/usr/bin/as"
	loader:= "/usr/bin/ld"
	ret := fmt.Sprintf("%s /tmp/%s -o /tmp/prog.o\n%s /tmp/prog.o -o /tmp/prog\n/tmp/prog", compiler, st.Name(),loader)
	for _, arg := range args {
		ret += " " + arg
	}
	return ret
}

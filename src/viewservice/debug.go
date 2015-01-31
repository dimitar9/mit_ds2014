package viewservice

import "fmt"
// Debugging
const Debug = 0

func DPrintf(format string, a ...interface{}) (n int, err error) {
if Debug > 0 {
n, err = fmt.Printf(format, a...)
}
return
}

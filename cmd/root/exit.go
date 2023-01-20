package root

import "os"

func ExitWithError() {
	// reserved exit codes: https://tldp.org/LDP/abs/html/exitcodes.html
	os.Exit(5)
}

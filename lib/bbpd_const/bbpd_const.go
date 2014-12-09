// Collection of global constant values.
package bbpd_const

const (
	INDENT        = "indent"
	COMPACT       = "compact"
	KEYS          = "keys"
	ATTRS         = "attrs"
	CONTENTTYPE   = "Content-Type"
	CONTENTLENGTH = "Content-Length"
	JSONMIME      = "application/json"
	PORT          = 12333 // primary port
	PORT2         = 12334 // secondary
	LOCALHOST     = "localhost"

	// request headers specific to bbpd
	X_BBPD_VERBOSE = "X-Bbpd-Verbose"
	X_BBPD_INDENT  = "X-Bbpd-Indent"
)

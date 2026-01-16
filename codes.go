package termite

const (
	termControlEraseLine     = "\033[K"
	termControlClearScreen   = "\033[H\033[2J"
	termControlCursorHide    = "\033[?25l"
	termControlCursorShow    = "\033[?25h"
	termControlCursorSave    = "\033[s"
	termControlCursorRestore = "\033[u"

	termControlCursorPositionFmt = "\033[%d;%dH"
	termControlCursorUpFmt       = "\033[%dA"
	termControlCursorDownFmt     = "\033[%dB"
	termControlCursorForwardFmt  = "\033[%dC"
	termControlCursorBackwardFmt = "\033[%dD"
)

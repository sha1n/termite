package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/sha1n/termite"
)

var taskDoneMarkUniChar = color.GreenString("\u2714")
var taskFailMarkUniChar = color.RedString("\u2717")
var progressPhases = []string{
	"Initializing...",
	"Configuring task...",
	"Starting...",
	"Running...",
	"Saving results...",
	"Cleaning up...",
	"Finishing...",
	taskDoneMarkUniChar,
}

const spinnerRefreshInterval = time.Millisecond * 50
const progressRefreshInterval = time.Millisecond * 100

const splash = `
 ____  ____  ____  _  _  __  ____  ____    ____  ____  _  _   __  
(_  _)(  __)(  _ \( \/ )(  )(_  _)(  __)  (    \(  __)( \/ ) /  \ 
  )(   ) _)  )   // \/ \ )(   )(   ) _)    ) D ( ) _) / \/ \(  O )
 (__) (____)(__\_)\_)(_/(__) (__) (____)  (____/(____)\_)(_/ \__/ 

`

type demoContext struct {
	out       io.Writer
	termWidth func() int
}

func main() {
	writer := termite.NewAutoFlushingWriter(os.Stdout)

	termite.Println(splash)

	demo(&demoContext{
		out: writer,
		termWidth: func() int {
			termWidth, _, _ := termite.GetTerminalDimensions()
			return termWidth
		},
	})
}

func demo(ctx *demoContext) {
	c := termite.NewCursor(ctx.out)
	c.Hide()
	defer c.Show()

	demoMatrix(ctx)
	demoSpinner(ctx)
	demoCursor(ctx)
	demoConcurrentProgressBars(ctx)
}

func demoMatrix(ctx *demoContext) {
	printTitle("Matrix Layout", ctx)

	m := termite.NewMatrix(termite.StdoutWriter, progressRefreshInterval)
	matrixCtx, cancel := context.WithCancel(context.Background())
	done := m.Start(matrixCtx)

	// allocating rows for 5 tasks and one space row
	m.NewRange(6)

	// adding a progress bar row
	progressRow := m.NewRow()
	pb := termite.NewProgressBar(
		progressRow,
		5*len(progressPhases),
		ctx.termWidth,
		ctx.termWidth()/8,
		termite.DefaultProgressBarFormatter(),
	)
	tick, _ := pb.Start(matrixCtx)

	update := func(rowIndex int, status string) {
		// to make it look more realistic we randomize task duration
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))

		row, _ := m.GetRow(rowIndex)
		row.Update(fmt.Sprintf("- Matrix Task %d - %s", rowIndex+1, status))
	}

	rand.Seed(time.Now().UnixNano())
	indexes := []int{0, 1, 2, 3, 4}
	// update the matrix
	for _, status := range progressPhases {
		// to make it look more realistic we update in a random order
		rand.Shuffle(len(indexes), func(i, j int) { indexes[i], indexes[j] = indexes[j], indexes[i] })
		for _, i := range indexes {
			update(i, status)
			tick("")
		}
	}

	cancel()
	<-done
	termite.Println("")
}

func demoSpinner(ctx *demoContext) {
	printTitle("Spinner progress indicator", ctx)

	m := termite.NewMatrix(termite.StdoutWriter, progressRefreshInterval)
	matrixCtx, cancel := context.WithCancel(context.Background())
	done := m.Start(matrixCtx)

	customFormatter1 := &customSpinnerFormatter{
		charSeq:           []string{"\u2588", "\u2587", "\u2586", "\u2585", "\u2584", "\u2583", "\u2582", "\u2581"},
		formatTitleFn:     color.CyanString,
		formatIndicatorFn: color.RedString,
	}
	customFormatter2 := &customSpinnerFormatter{
		charSeq:           []string{"/", "-", "\\", "|"},
		formatTitleFn:     color.MagentaString,
		formatIndicatorFn: color.GreenString,
	}
	spinners := []termite.Spinner{
		termite.NewSpinner(m.NewRow(), "Running...", spinnerRefreshInterval, termite.DefaultSpinnerFormatter()),
		termite.NewSpinner(m.NewRow(), "Running...", spinnerRefreshInterval, customFormatter1),
		termite.NewSpinner(m.NewRow(), "Running...", spinnerRefreshInterval, customFormatter2),
	}

	for _, spinner := range spinners {
		_ = spinner.Start(matrixCtx)
	}
	time.Sleep(time.Second)
	for _, spinner := range spinners {
		spinner.SetTitle("Finishing...")
	}
	time.Sleep(time.Second)
	for _, spinner := range spinners {
		_ = spinner.Stop("- Done " + taskDoneMarkUniChar)
	}

	cancel()
	<-done
	termite.Println("")
}

func demoCursor(ctx *demoContext) {
	printTitle("Cursor back tracking and line rewrites", ctx)

	fmtTaskStatus := func(name, status string) string {
		return fmt.Sprintf("- Task %s %s", name, status)
	}

	cursor := termite.NewCursor(termite.StdoutWriter)
	termite.Println(fmtTaskStatus("A", "pending..."))
	termite.Println(fmtTaskStatus("B", "pending..."))
	termite.Println(fmtTaskStatus("C", "pending..."))

	time.Sleep(time.Second * 1)
	cursor.Up(3)
	termite.Print(termite.TermControlEraseLine + fmtTaskStatus("A", taskDoneMarkUniChar))
	cursor.Down(3)

	time.Sleep(time.Millisecond * 50)
	cursor.Up(1)
	termite.Print(termite.TermControlEraseLine + fmtTaskStatus("C", taskDoneMarkUniChar))
	cursor.Down(1)

	time.Sleep(time.Millisecond * 50)
	cursor.Up(2)
	termite.Print(termite.TermControlEraseLine + fmtTaskStatus("B", taskFailMarkUniChar))
	cursor.Down(2)

	time.Sleep(time.Millisecond * 50)

	termite.Println("")
}

func demoConcurrentProgressBars(ctx *demoContext) {
	printTitle("Concurrent tasks progress", ctx)

	cursor := termite.NewCursor(termite.StdoutWriter)
	ticks := 200
	progressTickerWith := func(width int, formatter termite.ProgressBarFormatter) (func(), context.CancelFunc) {
		bar := termite.NewProgressBar(termite.StdoutWriter, ticks, ctx.termWidth, width, formatter)
		c, cancel := context.WithCancel(context.Background())
		tick, _ := bar.Start(c)
		actualTicks := 0

		return func() {
			actualTicks++
			tick(fmt.Sprintf("Running %d out of %d :", actualTicks, ticks))
			cursor.Down(1)
		}, cancel
	}

	var cancel1, cancel2, cancel3, cancel4 context.CancelFunc
	var tick1, tick2, tick3, tick4 func()

	termWidth := ctx.termWidth
	termite.AllocateNewLines(4) // allocate 4 lines
	tick1, cancel1 = progressTickerWith(termWidth()*3/16, &customProgressBarFormatter{Fill: '\u258C', formatBorderFn: color.WhiteString, formatFillFn: color.HiCyanString})
	tick2, cancel2 = progressTickerWith(termWidth()*1/4, &customProgressBarFormatter{Fill: '\u2592', formatBorderFn: color.YellowString, formatFillFn: color.BlueString})
	tick3, cancel3 = progressTickerWith(termWidth()*3/8, &customProgressBarFormatter{Fill: '\u2591', formatBorderFn: color.GreenString, formatFillFn: color.RedString})
	tick4, cancel4 = progressTickerWith(termWidth()*1/2, &customProgressBarFormatter{Fill: '\u2587', formatBorderFn: color.RedString, formatFillFn: color.GreenString})

	defer func() {
		cancel1()
		cancel2()
		cancel3()
		cancel4()
	}()
	tick := func() {
		tick1()
		tick2()
		tick3()
		tick4()
	}

	for i := 0; i < ticks; i++ {
		tick()
		time.Sleep(time.Millisecond * 10)
		cursor.Up(4)
	}
	cursor.Down(4)
	cursor.Show()

	termite.Println("\n")
}

func printTitle(s string, ctx *demoContext) {
	chars := len(s)
	border := strings.Repeat("-", chars+2)
	termite.Println(border)
	termite.Println(fmt.Sprintf(" %s ", color.GreenString(strings.Title(s))))
	termite.Println(border)
	termite.Println("")
}

type customSpinnerFormatter struct {
	charSeq           []string
	formatTitleFn     func(format string, a ...interface{}) string
	formatIndicatorFn func(format string, a ...interface{}) string
}

func (f *customSpinnerFormatter) FormatTitle(s string) string {
	return f.formatTitleFn(s)
}

func (f *customSpinnerFormatter) FormatIndicator(char string) string {
	return f.formatIndicatorFn(char)
}

func (f *customSpinnerFormatter) CharSeq() []string {
	return f.charSeq
}

type customProgressBarFormatter struct {
	Fill           rune
	formatBorderFn func(format string, a ...interface{}) string
	formatFillFn   func(format string, a ...interface{}) string
}

func (f *customProgressBarFormatter) FormatLeftBorder() string {
	return f.formatBorderFn(fmt.Sprintf("%c", termite.DefaultProgressBarLeftBorder))
}

func (f *customProgressBarFormatter) FormatRightBorder() string {
	return f.formatBorderFn(fmt.Sprintf("%c", termite.DefaultProgressBarLeftBorder))
}

func (f *customProgressBarFormatter) FormatFill() string {
	return f.formatFillFn(fmt.Sprintf("%c", f.Fill))
}

func (f *customProgressBarFormatter) FormatBlank() string {
	return f.formatFillFn(fmt.Sprintf("%c", termite.DefaultProgressBarBlank))
}

func (f *customProgressBarFormatter) MessageAreaWidth() int {
	return 25
}

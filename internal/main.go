package main

import (
	"context"
	"fmt"
	"math/rand"
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
const progressRefreshInterval = time.Millisecond * 10

const splash = `
 ____  ____  ____  _  _  __  ____  ____    ____  ____  _  _   __  
(_  _)(  __)(  _ \( \/ )(  )(_  _)(  __)  (    \(  __)( \/ ) /  \ 
  )(   ) _)  )   // \/ \ )(   )(   ) _)    ) D ( ) _) / \/ \(  O )
 (__) (____)(__\_)\_)(_/(__) (__) (____)  (____/(____)\_)(_/ \__/ 

`

func main() {
	t := termite.NewTerminal()
	t.Println(splash)
	demo(t)
}

func demo(t termite.Terminal) {
	c := termite.NewCursor(termite.StdoutWriter)
	c.Hide()
	defer c.Show()

	demoMatrix(t)
	demoSpinner(t)
	demoCursor(t)
	demoConcurrentProgressBars(t)
}

func demoMatrix(t termite.Terminal) {
	printTitle("Matrix Layout", t)

	m := termite.NewMatrix(termite.StdoutWriter, progressRefreshInterval)
	cancel := m.Start()

	// allocating rows for 5 tasks and one space row
	m.NewRange(6)

	// adding a progress bar row
	progressRow := m.NewRow()
	pb := termite.NewProgressBar(progressRow, 5*len(progressPhases), t.Width(), t.Width()/8, '\u2587', '\u2587', '\u2587')
	tick, _, _ := pb.Start()

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
			tick()
		}
	}

	cancel()
	t.Println("")
}

func demoSpinner(t termite.Terminal) {
	printTitle("Spinner progress indicator", t)

	spinner := termite.NewSpinner(termite.StdoutWriter, "Running...", spinnerRefreshInterval)
	if _, e := spinner.Start(); e == nil {
		time.Sleep(time.Second * 1)
		spinner.SetTitle("Finishing...")
		time.Sleep(time.Second * 1)
		_ = spinner.Stop("- Done " + taskDoneMarkUniChar)
	}

	t.Println("\r\n")
}

func demoCursor(t termite.Terminal) {
	printTitle("Cursor back tracking and line rewrites", t)

	fmtTaskStatus := func(name, status string) string {
		return fmt.Sprintf("- Task %s %s", name, status)
	}

	cursor := termite.NewCursor(termite.StdoutWriter)
	t.Println(fmtTaskStatus("A", "pending..."))
	t.Println(fmtTaskStatus("B", "pending..."))
	t.Println(fmtTaskStatus("C", "pending..."))

	time.Sleep(time.Second * 1)
	cursor.Up(3)
	t.Print(termite.TermControlEraseLine + fmtTaskStatus("A", taskDoneMarkUniChar))
	cursor.Down(3)

	time.Sleep(time.Millisecond * 50)
	cursor.Up(1)
	t.Print(termite.TermControlEraseLine + fmtTaskStatus("C", taskDoneMarkUniChar))
	cursor.Down(1)

	time.Sleep(time.Millisecond * 50)
	cursor.Up(2)
	t.Print(termite.TermControlEraseLine + fmtTaskStatus("B", taskFailMarkUniChar))
	cursor.Down(2)

	time.Sleep(time.Millisecond * 50)

	t.Println("")
}

func demoConcurrentProgressBars(t termite.Terminal) {
	printTitle("Concurrent tasks progress", t)

	cursor := termite.NewCursor(termite.StdoutWriter)
	ticks := 20
	progressTickerWith := func(width int, fill rune) (func(), context.CancelFunc) {
		bar := termite.NewProgressBar(termite.StdoutWriter, ticks, width, t.Width(), fill, fill, fill)
		tick, cancel, _ := bar.Start()

		return func() {
			tick()
			cursor.Down(1)
		}, cancel
	}

	var cancel1, cancel2, cancel3, cancel4 context.CancelFunc
	var tick1, tick2, tick3, tick4 func()

	t.AllocateNewLines(4) // allocate 4 lines
	tick1, cancel1 = progressTickerWith(t.Width()*1/8, '\u258C')
	tick2, cancel2 = progressTickerWith(t.Width()*1/4, '\u2592')
	tick3, cancel3 = progressTickerWith(t.Width()*3/8, '\u2591')
	tick4, cancel4 = progressTickerWith(t.Width()*1/2, '\u2587')

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

	for i := 0; i < 20; i++ {
		tick()
		time.Sleep(time.Millisecond * 10)
		cursor.Up(4)
	}
	cursor.Down(4)
	cursor.Show()

	t.Println("\n")
}

func printTitle(s string, t termite.Terminal) {
	chars := len(s)
	border := strings.Repeat("-", chars+2)
	t.Println(border)
	t.Println(fmt.Sprintf(" %s ", color.GreenString(strings.Title(s))))
	t.Println(border)
	t.Println("")
}

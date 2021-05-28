package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/sha1n/termite"
)

var taskDoneMarkUniChar = color.GreenString("\u2714")
var taskFailMarkUniChar = color.RedString("\u2717")

const splash = `
 ____  ____  ____  _  _  __  ____  ____    ____  ____  _  _   __  
(_  _)(  __)(  _ \( \/ )(  )(_  _)(  __)  (    \(  __)( \/ ) /  \ 
  )(   ) _)  )   // \/ \ )(   )(   ) _)    ) D ( ) _) / \/ \(  O )
 (__) (____)(__\_)\_)(_/(__) (__) (____)  (____/(____)\_)(_/ \__/ 

`

func main() {
	t := termite.NewTerminal(true)
	t.Println(splash)
	demo(t)
}

func demo(t termite.Terminal) {
	c := termite.NewCursor(t)
	c.Hide()
	defer c.Show()

	demoMatrix(t)
	demoSpinner(t)
	demoCursor(t)
	demoProgressBars(t)
	demoConcurrentProgressBars(t)
}

func demoMatrix(t termite.Terminal) {
	printTitle("Matrix Layout", t)

	refreshInterval := time.Millisecond * 10
	m := termite.NewMatrix(t, refreshInterval)
	cancel := m.Start()

	lines := []io.StringWriter{
		m.NewLineStringWriter(), m.NewLineStringWriter(), m.NewLineStringWriter(), m.NewLineStringWriter(), m.NewLineStringWriter(),
	}

	for i := 0; i < 100; i++ {
		time.Sleep(refreshInterval)
		_, _ = lines[i%len(lines)].WriteString(fmt.Sprintf("- Matrix Line -> version %d", i+1))
	}

	cancel()
	t.Println("")
}

func printTitle(s string, t termite.Terminal) {
	chars := len(s)
	border := strings.Repeat("-", chars+2)
	t.Println(border)
	t.Println(fmt.Sprintf(" %s ", color.GreenString(strings.Title(s))))
	t.Println(border)
	t.Println("")
}

func demoSpinner(t termite.Terminal) {
	printTitle("Spinner progress indicator", t)

	spinner := termite.NewSpinner(t, "Running...", 100)
	if _, e := spinner.Start(); e == nil {
		time.Sleep(time.Second * 1)
		spinner.SetTitle("Finishing...")
		time.Sleep(time.Second * 1)
		_ = spinner.Stop("- Done " + taskDoneMarkUniChar)
		t.Println("")
	}
}

func demoCursor(t termite.Terminal) {
	printTitle("Cursor back tracking and line rewrites", t)

	fmtTaskStatus := func(name, status string) string {
		return fmt.Sprintf("- Task %s %s", name, status)
	}

	cursor := termite.NewCursor(t)
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

func demoProgressBars(t termite.Terminal) {
	printTitle("Default progress bar", t)

	pb := termite.NewDefaultProgressBar(t, 20)
	for pb.Tick() {
		time.Sleep(time.Millisecond * 10)
	}

	t.Println("\n")
}

func demoConcurrentProgressBars(t termite.Terminal) {
	printTitle("Concurrent tasks progress", t)

	cursor := termite.NewCursor(t)

	t.AllocateNewLines(4) // allocate 4 lines

	b1 := termite.NewProgressBar(t, 1000, t.Width()*1/8, '\u258C', '\u258C', '\u258C')
	b2 := termite.NewProgressBar(t, 1000, t.Width()*1/4, '\u258F', '\u2595', '\u2592')
	b3 := termite.NewProgressBar(t, 1000, t.Width()*3/8, '\u258F', '\u2595', '\u2591')
	b4 := termite.NewProgressBar(t, 1000, t.Width()*1/2, '\u2587', '\u2587', '\u2587')

	t1, _, _ := b1.Start()
	t2, _, _ := b2.Start()
	t3, _, _ := b3.Start()
	t4, _, _ := b4.Start()

	for i := 0; i < 1000; i++ {
		t1()
		cursor.Down(1)
		t2()
		cursor.Down(1)
		t3()
		cursor.Down(1)
		t4()
		time.Sleep(1 * time.Millisecond)
		cursor.Up(3)
	}
	cursor.Down(3)
	cursor.Show()

	t.Println("\n")
}

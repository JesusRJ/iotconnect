// Simple program that displays the state of the specified joystick
//
//     go run joysticktest.go 2
// displays state of joystick id 2
package main

import (
  "fmt"
  "github.com/nsf/termbox-go"
  "github.com/simulatedsimian/joystick"
  "os"
  "strconv"
  "time"
)

func printAt(x, y int, s string) {
  for _, r := range s {
    termbox.SetCell(x, y, r, termbox.ColorDefault, termbox.ColorDefault)
    x++
  }
}

func translateAxis(data int) (pos int) {
  switch {
  case data <= -32767:
    pos = 1
  case data <= -16384:
    pos = 2
  case data == 0:
    pos = 3
  case data <= 16384:
    pos = 4
  case data <= 32768:
    pos = 5
  }
  return
}

func readJoystick(js joystick.Joystick) {
  jinfo, err := js.Read()

  if err != nil {
    printAt(1, 5, "Error: "+err.Error())
    return
  }

  printAt(1, 5, "Buttons:")
  for button := 0; button < js.ButtonCount(); button++ {
    if jinfo.Buttons&(1<<uint32(button)) != 0 {
      printAt(10+button, 5, "X")
    } else {
      printAt(10+button, 5, ".")
    }
  }

  for axis := 0; axis < js.AxisCount(); axis++ {
    printAt(1, axis+7, fmt.Sprintf("Axis %2d Value: %7d", axis, jinfo.AxisData[axis]))
  }

  pos := 8 + js.AxisCount()
  // Clean
  for x := 1; x < 6; x++ {
    printAt(1, x+pos, "|      |")
  }

  x := translateAxis(jinfo.AxisData[0])
  y := translateAxis(jinfo.AxisData[1])
  termbox.SetCell(x+1, y+pos, 'x', termbox.ColorGreen, termbox.ColorDefault)

  return
}

func main() {

  jsid := 0
  if len(os.Args) > 1 {
    i, err := strconv.Atoi(os.Args[1])
    if err != nil {
      fmt.Println(err)
      return
    }
    jsid = i
  }

  js, jserr := joystick.Open(jsid)

  if jserr != nil {
    fmt.Println(jserr)
    return
  }

  err := termbox.Init()
  if err != nil {
    panic(err)
  }
  defer termbox.Close()

  eventQueue := make(chan termbox.Event)
  go func() {
    for {
      eventQueue <- termbox.PollEvent()
    }
  }()

  ticker := time.NewTicker(time.Millisecond * 40)

  for doQuit := false; !doQuit; {
    select {
    case ev := <-eventQueue:
      if ev.Type == termbox.EventKey {
        if ev.Ch == 'q' {
          doQuit = true
        }
      }
      if ev.Type == termbox.EventResize {
        termbox.Flush()
      }

    case <-ticker.C:
      printAt(1, 0, "-- Press 'q' to Exit --")
      printAt(1, 1, fmt.Sprintf("Joystick Name: %s", js.Name()))
      printAt(1, 2, fmt.Sprintf("   Axis Count: %d", js.AxisCount()))
      printAt(1, 3, fmt.Sprintf(" Button Count: %d", js.ButtonCount()))
      readJoystick(js)
      termbox.Flush()
    }
  }
}

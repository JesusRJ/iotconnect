// Simple program that displays the state of the specified joystick
//
//     go run joysticktest.go 2
// displays state of joystick id 2
package main

import (
  "fmt"
  "github.com/nsf/termbox-go"
  "github.com/simulatedsimian/joystick"
  "log"
  "net"
  "os"
  "strconv"
  "time"
)

// UDP address
var (
  RemoteAddr *net.UDPAddr
  RemoteConn *net.UDPConn
  CarState   int
)

// Connect to car: send message by broadcast and
// get car's IP address
func connectcar() {
  laddr, err := net.ResolveUDPAddr("udp", ":0")
  if err != nil {
    log.Fatal(err)
  }

  maddr, err := net.ResolveUDPAddr("udp4", "255.255.255.255:8089")
  if err != nil {
    log.Fatal(err)
  }

  c, err := net.ListenUDP("udp4", laddr)
  if err != nil {
    log.Fatal(err)
  }

  defer c.Close()

  for RemoteAddr == nil {
    _, err := c.WriteToUDP([]byte("cmd=ping"), maddr)

    if err != nil {
      log.Fatal(err)
    }

    // Set timeout to 10s
    c.SetReadDeadline(time.Now().Add(10 * time.Second))

    buf := make([]byte, 1024)
    if _, addr, err := c.ReadFromUDP(buf); err != nil {
      log.Println(err)
      RemoteAddr = nil
    } else {
      RemoteAddr = addr
      RemoteConn, err = net.DialUDP("udp", nil, RemoteAddr)
      if err != nil {
        log.Fatal(err)
      }
    }
  }
}

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

func readJoystick(js joystick.Joystick) (x, y int) {
  jinfo, err := js.Read()

  if err != nil {
    printAt(1, 6, "Error: "+err.Error())
    return
  }

  printAt(1, 6, "Buttons:")
  for button := 0; button < js.ButtonCount(); button++ {
    if jinfo.Buttons&(1<<uint32(button)) != 0 {
      printAt(10+button, 6, "X")
    } else {
      printAt(10+button, 6, ".")
    }
  }

  for axis := 0; axis < js.AxisCount(); axis++ {
    printAt(1, axis+8, fmt.Sprintf("Axis %2d Value: %7d", axis, jinfo.AxisData[axis]))
  }

  pos := 9 + js.AxisCount()
  // Clean
  for x := 1; x < 6; x++ {
    printAt(1, x+pos, "|      |")
  }

  x = translateAxis(jinfo.AxisData[0])
  y = translateAxis(jinfo.AxisData[1])
  termbox.SetCell(x+1, y+pos, 'x', termbox.ColorGreen, termbox.ColorDefault)

  return
}

func sendCommand(x, y int) string {
  var state int

  switch {
  case y == 1:
    state = 1
  case y == 5:
    state = 2
  case x == 1:
    state = 4
  case x == 5:
    state = 3
  default:
    state = 0
  }

  if state == CarState {
    return ""
  } else {
    CarState = state
  }

  // write a message to server
  message := []byte(fmt.Sprintf("cmd=control&d=%d", state))

  RemoteConn.SetWriteDeadline(time.Now().Add(5 * time.Second))

  _, err := RemoteConn.Write(message)

  if err != nil {
    log.Println(err)
    // printAt(30, 21, fmt.Sprintf("%s", err))
    RemoteAddr = nil
    RemoteConn = nil
    // Try reconnect car
    connectcar()
  }

  // receive message from server
  buffer := make([]byte, 1024)
  n, _, err := RemoteConn.ReadFromUDP(buffer)
  return string(buffer[:n])
}

func main() {
  // Get remote address
  connectcar()

  // Get Joystick reference
  jsid := 0
  if len(os.Args) > 1 {
    i, err := strconv.Atoi(os.Args[1])
    if err != nil {
      log.Fatal(err)
      return
    }
    jsid = i
  }

  js, jserr := joystick.Open(jsid)

  if jserr != nil {
    fmt.Println(jserr)
    return
  }

  // Init termbox
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

  // Process
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
      printAt(1, 4, fmt.Sprintf("  UDP address: %s \n", RemoteConn.RemoteAddr().String()))

      x, y := readJoystick(js)
      printAt(30, 19, fmt.Sprintf("X: %d, Y: %d", x, y))
      msg := sendCommand(x, y)
      printAt(30, 20, fmt.Sprintf("%s", msg))
      termbox.Flush()
    }
  }
}

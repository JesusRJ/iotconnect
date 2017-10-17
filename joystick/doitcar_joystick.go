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

var carState int

// Connect to car: send message by broadcast and
// get car's IP address
func connectCar() *net.UDPConn {
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

  var udpconn *net.UDPConn

  for udpconn == nil {
    _, err := c.WriteToUDP([]byte("cmd=ping"), maddr)

    if err != nil {
      log.Fatal(err)
    }

    // Set timeout to 10s
    c.SetReadDeadline(time.Now().Add(3 * time.Second))

    buf := make([]byte, 1024)
    if _, addr, err := c.ReadFromUDP(buf); err != nil {
      log.Println(err)
      udpconn = nil
    } else {
      udpconn, err = net.DialUDP("udp", nil, addr)
      if err != nil {
        log.Fatal(err)
      }
    }
  }

  return udpconn
}

func connectJoystick(id int) joystick.Joystick {
  var js joystick.Joystick = nil
  var jserr error

  for js == nil {
    js, jserr = joystick.Open(id)

    if jserr != nil {
      printAt(1, 22, "Error: "+jserr.Error())
      time.Sleep(5 * time.Second)
    }
  }
  return js
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

func readJoystick(js joystick.Joystick) (x, y int, e error) {
  jinfo, err := js.Read()

  if err != nil {
    printAt(1, 6, "Error: "+err.Error())
    return 0, 0, err
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

func sendCommand(x, y int, remoteConn *net.UDPConn) (string, error) {
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

  if state == carState {
    return "", nil
  } else {
    carState = state
  }

  // write a message to server
  message := []byte(fmt.Sprintf("cmd=control&d=%d", state))

  remoteConn.SetWriteDeadline(time.Now().Add(5 * time.Second))

  _, err := remoteConn.Write(message)

  if err != nil {
    log.Println(err)
    return "", err
  }

  // receive message from server
  buffer := make([]byte, 1024)
  n, _, err := remoteConn.ReadFromUDP(buffer)
  return string(buffer[:n]), nil
}

func main() {
  // Get remote address
  remoteConn := connectCar()

  defer remoteConn.Close()

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

  js := connectJoystick(jsid)

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
      printAt(1, 4, fmt.Sprintf("  UDP address: %s\n", remoteConn.RemoteAddr().String()))

      x, y, err := readJoystick(js)
      if err != nil {
        log.Println(err)
        js = connectJoystick(jsid)
      }

      printAt(30, 19, fmt.Sprintf("X: %d, Y: %d", x, y))
      msg, _ := sendCommand(x, y, remoteConn)
      printAt(30, 20, fmt.Sprintf("%s", msg))
      termbox.Flush()
    }
  }
}

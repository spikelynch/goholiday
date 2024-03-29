package main

import (
    "fmt"
    "net"
    "os"
    "bytes"
    "encoding/binary"
    "time"
    "math/rand"
    "github.com/lucasb-eyer/go-colorful"
)


const R = 11.0
const H = 12.0

const NLIGHTS = 25
const VMAX = 5

const WAIT = 400
const PAUSE = 8000

const FADE_STEPS = 40
const FADE_WAIT = 40

const HISTCAP = 1000

const ONSAT = 1
const ONVAL = 0.8
const ONVALRANGE = 0.2
const OFFSAT = 0.7
const OFFVAL = 0.3
const OFFVALRANGE = 0.3
const RANDCOMP = 60.0

const X_SIZE = 10
const Y_SIZE = 5

type Holiday struct {
	Header [10]uint8
	Globes [150]uint8
}





// func makeMap() [][]int {
//     m := make([][]int, 8)

//     for i := 0; i < 8; i++ {
//         m[i] = make([]int, 6)
//         m[i][0] = i;
//         m[i][1] = 15 - i;
//         m[i][2] = 17 + i;
//         m[i][3] = 32 - i;
//         m[i][4] = 34 + i;
//         m[i][5] = 49 - i;
//     }

//     return m

// }

func makeMap() [][]int {
    m := make([][]int, 5)

    for i := 0; i < 5; i++ {
        m[i] = make([]int, 10)
        m[i][0] = i;
        m[i][1] = 9 - i;
        m[i][2] = 10 + i;
        m[i][3] = 19 - i;
        m[i][4] = 20 + i;
        m[i][5] = 29 - i;
        m[i][6] = 30 + i;
        m[i][7] = 39 - i;
        m[i][8] = 40 + i;
        m[i][9] = 49 - i;
        fmt.Printf("%d: %v\n", i, m[i]);
    }

    return m

}


func showHoliday(hol *Holiday) {
    for i := 0; i < 50; i++ {
        fmt.Printf("hol %d = %d %d %d\n", i, hol.Globes[i * 3], hol.Globes[i * 3 + 1], hol.Globes[i * 3 + 2])
    }
}

func setHolidayGlobe(hol *Holiday, i int, c colorful.Color) {
    hol.Globes[i * 3] = uint8(63 * c.R)
    hol.Globes[i * 3 + 1] = uint8(63 * c.G)
    hol.Globes[i * 3 + 2] = uint8(63 * c.B)
}


func sendHoliday(conn *net.UDPConn, hol *Holiday) {
    datagram := new(bytes.Buffer)
    enc_err := binary.Write(datagram, binary.LittleEndian, hol)
    if enc_err != nil {
        fmt.Println("Encoding failed: ", enc_err)
        return
    }

    _, err := conn.Write(datagram.Bytes())

    if err != nil {
        fmt.Println(err)
    }
}


func wrap(i, max int) int {
    if i < 0 {
        return i + max
    } else if i > max - 1 {
        return i - max
    }
    return i
}


func neighbours(x, y, xs, ys int, board [][]bool) int {
    count := 0
    for i := -1; i < 2; i++ {
        for j := -1; j < 2; j++ {
            if !(i == 0 && j == 0) {
                xn := wrap(x + i, xs)
                yn := wrap(y + j, ys)
                if board[xn][yn] {
                    count++
                }
            }
        }
    }
    return count
}


func gameoflife(board [][]bool) [][]bool {
    xs := len(board)
    ys := len(board[0])
    nboard := make([][]bool, xs)
    for x := 0; x < xs; x++ {
        nboard[x] = make([]bool, ys)
        for y := 0; y < ys; y++ {
            n := neighbours(x, y, xs, ys, board)
            if( board[x][y] ) {
                if n < 2 || n > 3 {
                    nboard[x][y] = false
                } else {
                    nboard[x][y] = true
                }
            } else {
                if n == 3 {
                    nboard[x][y] = true
                } else {
                    nboard[x][y] = false
                }
            }
        }
    }
    return nboard
}

func initBoard(xs, ys int) [][]bool {
    board := make([][]bool, xs)
    for x := 0; x < xs; x++ {
        board[x] = make([]bool, ys)
        for y := 0; y < ys; y++ {
            board[x][y] = ( rand.Float64() < 0.4 )
        }
    }
    return board
}




func serialiseBoard(board [][]bool) uint64 {
    var s uint64
    var e uint64
    e = 1
    for _, row := range board {
        for _, cell := range row {
            if cell {
                s += e
            }
            e = e * 2
        }
    }
    return s
}

func histContains(h []uint64, n uint64) bool {
    for _, h1 := range h {
        if h1 == n {
            return true
        }
    }
    return false
}

func copyBoard(board [][]bool) [][]bool {
    cb := make([][]bool, len(board))
    for i, row := range board {
        cb[i] = make([]bool, len(row))
        copy(cb[i], row)
    }
    return cb
}

func printBoard(board [][]bool) {
    for _, row := range board {
        for _, col := range row {
            if col {
                fmt.Printf("[*]")
            } else {
                fmt.Printf("[ ]")
            }
        }
        fmt.Printf("\n")
    }
    fmt.Printf("---\n")
}


func colorPair() []colorful.Color {
    hueOn := rand.Float64() * 360
    hueOff := hueOn + 180 - (rand.Float64() * 2 * RANDCOMP - RANDCOMP)
    if hueOff > 360 {
        hueOff -= 360
    } else if hueOff < 0 {
        hueOff += 360
    }
    pair := make([]colorful.Color, 2)
    pair[0] = colorful.Hsv(hueOff, OFFSAT, OFFVAL + OFFVALRANGE * rand.Float64())
    pair[1] = colorful.Hsv(hueOn, ONSAT, ONVAL + ONVALRANGE * rand.Float64())
    return pair
}

func makeGradient(steps int) []colorful.Color {
    p := colorPair()
    g := make([]colorful.Color, steps)
    for k := 0; k < steps; k++ {
        g[k] = p[0].BlendHcl(p[1], float64(k) / float64(steps - 1)).Clamped()
    }
    return g
}


func animateBoard(m [][]int, b1, b2 [][]bool, gradient []colorful.Color, conn *net.UDPConn, hol *Holiday) {
    on := gradient[len(gradient) - 1]
    off := gradient[0]
    for k, _ := range gradient {
        offward := gradient[len(gradient) - 1 - k]
        for i, row := range m {
            for j, globe := range row {
                if b1[i][j] == b2[i][j] {
                    if b2[i][j] {
                        setHolidayGlobe(hol, globe, on)
                    } else {
                        setHolidayGlobe(hol, globe, off)
                    }
                } else {
                    if b2[i][j] {
                        setHolidayGlobe(hol, globe, on)
                    } else {
                        setHolidayGlobe(hol, globe, offward)
                    }
                }
            }
        }
        sendHoliday(conn, hol)
        time.Sleep(FADE_WAIT * time.Millisecond)
    }
}




func main() {
    arguments := os.Args
    if len(arguments) == 1 {
        fmt.Println("Please provide a host:port string")
        return
    }
    CONNECT := arguments[1]

    s, err := net.ResolveUDPAddr("udp4", CONNECT)
    c, err := net.DialUDP("udp4", nil, s)
    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Printf("Holiday is %s\n", c.RemoteAddr().String())
    defer c.Close()

    rand.Seed(time.Now().UnixNano())

    hol := new(Holiday)

    m := makeMap()

    gradient := makeGradient(FADE_STEPS)

    board := initBoard(Y_SIZE, X_SIZE)

    tick := 0

    history := make([]uint64, 0, HISTCAP)

    for {
        oboard := board
        board = gameoflife(board)

        animateBoard(m, oboard, board, gradient, c, hol)

        // for i, row := range m {
        //     for j, globe := range row {
        //         if board[i][j] {
        //             setHolidayGlobe(hol, globe, palette[1])
        //         } else {
        //             setHolidayGlobe(hol, globe, palette[0])
        //         }
        //     }
        // }
        // sendHoliday(c, hol)

        time.Sleep(WAIT * time.Millisecond)
        tick++

        sb := serialiseBoard(board)

        if histContains(history, sb) || tick > HISTCAP * 10 {
            board = initBoard(Y_SIZE, X_SIZE)
            gradient = makeGradient(FADE_STEPS)
            history = make([]uint64, 0, HISTCAP)
	    time.Sleep(PAUSE * time.Millisecond)
            tick = 0
        } else {
            history = append(history, sb)
        }
    }
}

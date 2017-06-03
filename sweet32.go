package main

import (
    "os"
    "fmt"
    "net"
    "time"
    "flag"
    "crypto/tls"
)

var verbose bool

func vprint(msg string) {
    if verbose {
        fmt.Printf(msg)
    } 
}


func banner() {
    fmt.Println("            [      Sweet Tea     ]            ")
    fmt.Println("            [ The SWEET32 Tester ]            ")
}

func check(e error) {
    if e != nil {
        fmt.Println(e)
        os.Exit(1)
    }
}


func cipherstring(i uint16) string {
    switch {
    case i == 0x000a:
        return "TLS_RSA_WITH_3DES_EDE_CBC_SHA"
    case i == 0xc012:
        return "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA"
    default:
        return ""
    }
}


func getConnection(server string, conf *tls.Config, timeout time.Duration) (*tls.Conn) {
    // Create a TCP connection.
    conn, err := net.DialTimeout("tcp", server, timeout)
    check(err)
    vprint(fmt.Sprintf("[+] Successfully connected to: %s\n", conn.RemoteAddr()))

    // Create TLS connection using our TCP connection and set a deadline
    // before attempting the handshake. This will ensure the handshake times
    // out.
    tlsconn := tls.Client(conn, conf)
    tlsconn.SetDeadline(time.Now().Add(timeout))

    err = tlsconn.Handshake()
    if err != nil {
        fmt.Println("[-] Unable to complete TLS handshake.")
        os.Exit(0)
    }

    // Reset the deadline to zero.
    tlsconn.SetDeadline(time.Time{})

    // Document cipher suite
    state := tlsconn.ConnectionState()
    vprint(fmt.Sprintf("[+] Using: %s\n", cipherstring(state.CipherSuite)))

    return tlsconn
}


func main() {
    var host string
    var port string

    flag.BoolVar(&verbose, "v", false, "Verbose output.")
    flag.StringVar(&host, "h", "", "IP address or hostname of web server.")
    flag.StringVar(&port, "p", "", "Port number of web server.")

    flag.Parse()

    if host == "" || port == "" {
        flag.Usage()
        os.Exit(0)
    }

    server := fmt.Sprintf("%s:%s", host, port)
    timeout := 30 * time.Second

    banner()
    fmt.Printf("[*] Testing connection to %s.\n", server)

    // Build TLS Config
    conf := &tls.Config{
        InsecureSkipVerify: true,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
            tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
        },
    }

    // Make our connection

    conn := getConnection(server, conf, timeout)
    defer conn.Close()

    // Write data to the connection.
    for i := 1; i <= 10000; i++ {
        send := []byte(fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\n\r\n", server))
        _, err := conn.Write(send)
        if err != nil {
            vprint("\n")
            fmt.Printf("[+] Connection closed after %d requests. Server is not vulnerable.\n", i)
            break
        }

        resp := make([]byte, 512)
        conn.Read(resp)

        if i % 20 == 0 && verbose {
            fmt.Printf(".")
        }

        if i == 10000 {
            fmt.Println("\n")
            fmt.Println("[-] The server accepted 10000 requests. Server is likely vulnerable.")
        }
    }
    fmt.Println("")
}

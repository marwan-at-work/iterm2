# iTerm2

### Go library for automating iTerm2 Scripts

### Install

go get marwan.io/iterm2

### Usage

```golang
package main

func main() {
    app, err := iterm2.NewApp("MyCoolPlugin")
    handle(err)
    defer app.Close()
    // use app to create or list windows, tabs, and sessions and send various commands to the terminal.
}
```

### How do I actually run the script?

- Since you will be using this library in a "main" program, you can literally just run the Go program through "go run" or install your program/binary globally through "go install" and then run it from any terminal.

- A nicer way to run the script is to "register" the plugin with iTerm2 so you can run it from iTerm's command pallette (cmd+shift+o). This means you won't need a terminal tab open or to remember what the plugin name is. See the following section on how to do that:

- Ensure you enable the Python API: https://iterm2.com/python-api-auth.html

### Installing the plugin into iTerm2

1. Install your program into your local PATH (for example running `go install`)
2. `go get marwan.io/iterm2/cmd/goiterm`
3. `goiterm install <bin>`
4. From any iTerm window run "cmd+shift+o" and look for `<bin>.py`.

### Progress

This is currently a work in progress and it is a subset of what the iTerm2 WebSocket protocol provides.

I don't intend to implement all of it as I am mainly implementing only the parts that I need for daily work. 

If you'd like to add more features, feel free to open an issue and a PR. 

# iTerm2

### Go library for automating iTerm2 Scripts

### Install

go get marwan.io/iterm2

### Usage

```golang
app, err := iterm2.NewApp()
handle(err)
defer app.Close()
// use app to create or list windows, tabs, and sessions and send various commands to the terminal.
```

### Progress

This is currently a work in progress and it is a subset of what the iTerm2 WebSocket protocol provides.

I don't intend to implement all of it as I am mainly implementing only the parts that I need for daily work. 

If you'd like to add more features, feel free to open an issue and a PR. 

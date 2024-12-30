# Chat App (WebRTC + Golang + BubbleTea)


## Introduction

- This project uses [BubbleTea](https://github.com/charmbracelet/bubbletea) for the UI.
- You can install it with `go get github.com/charmbracelet/bubbletea`.


## Setup

- You need to run `go mod tidy` to install the dependencies.
- Run `go run ./backend` to start the backend signalling server.
- Run `go run ./cmd/main.go` to start the room and chat.

## Usage

> Run Backend

![Backend](./assets/backend.png)

> Start chat

![Chat Init](./assets/chat-init.png)

> Host room

![Host Room](./assets/room-hosting.png)

> Room welcome

![Room Welcome](./assets/room-welcome.png)

> Peer joining

![Peer Joining](./assets/joining-room.png)

> Peer configuration

![Peer Joining](./assets/peer-setup.png)

> Peer window

![Peer Window](./assets/peer-joining.png)

> Chat Window (with history)

![Chat Window](./assets/chatting.png)
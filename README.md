# Wish Background Padding Minimal

Project that minimally reproduces potential bug in [Bubble Tea](https://github.com/charmbracelet/bubbletea) [Wish](https://github.com/charmbracelet/wish) or [Lipgloss](https://github.com/charmbracelet/lipgloss) (not sure which one) - background color is not applied to padding when serving the application from Docker (seems to work as expected during dev)

## Step to reproduce bug

- Go version 1.24.2

1. local

run application

```console
go run .
```

connect to TUI

```shell
ssh localhost -p 8080
```

notice that the background color is set for padding as expected

![local dev](./docs/local.png)

2. docker

build and run container

```shell
docker build -t wish-padding-test:0.0.1 .
```

```shell
docker run -p "8080:8080" -e "PORT=8080" -e "HOST=0.0.0.0" wish-padding-test:0.0.1
```

connect to TUI

```shell
ssh localhost -p 8080
```

notice that the background color is not applied to padding

![docker deployment](./docs/docker.png)

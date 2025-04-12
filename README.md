# sesh

`sesh` is a minimal CLI tool to declaratively launch tmux sessions from a YAML spec.

## Example YAML specification

```yaml
session: example-session
windows:
  - name: edit
    layout: tiled
    panes:
      - command: nvim
        path: ~/projects/project/
      - command: git status
        path: ~/projects/project/

  - name: serve
    layout: even-horizontal
    panes:
      - command: ./run-server
        path: ~/projects/project/
      - command: tail -f logs/server.log
        path: ~/projects/project/

  - name: misc
    panes:
      - command: htop
      - command: bash
```

## Development

```sh
go build -o build/sesh && ./build/sesh session-spec.yaml --verbose
```

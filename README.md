# podswap

PodSwap is a lightweight application designed to update your containers with almost no downtime. By listening for events in your GitHub repository, PodSwap re-builds and swaps your containers automatically.

This project is in its early stages. Feel free to give feedback and contribute.


## Installation

```bash
go install github.com/douglascdev/podswap@latest
```

If `GOPATH` and `GOBIN` are not set, it goes to `$HOME/go/bin` on UNIX systems,
which you can add to your `PATH` with:

```bash
export PATH=$PATH:$HOME/go/bin
```

## Usage

```bash
Usage of podswap:
  -f value
      path for the yml action file(default: working directory/.github/workflows/podswap.yml)
  -p value
      root path for the project(default: working directory)
```

Step-by-step setup:
- Set up an [ngrok account](https://dashboard.ngrok.com/signup) and get [your token](https://dashboard.ngrok.com/get-started/your-authtoken).
- Create a `.github/workflows/podswap.yml` file with the contents:

```yml
name: podswap

on:
  push:
    branches:
      - 'main'

jobs:
  podswap:
    uses: douglascdev/podswap/.github/workflows/action.yml
    with:
      pre-build-cmd: 'git pull'
      build-cmd: 'podman compose build'
      deploy-cmd: 'podman compose up -d --force-recreate'
    secrets:
      WEBHOOK_SECRET: ${{ secrets.WEBHOOK_SECRET }}
      WEBHOOK_URL: ${{ secrets.WEBHOOK_URL }}
```

- Create 2 secrets on your project's repo in  `Settings > Secrets and variables > Actions > New secret`:
  - `WEBHOOK_SECRET`: create a good secret(pretend it's a password).
  - `WEBHOOK_URL`: paste the URL from [domains](https://dashboard.ngrok.com/domains).

- Run the listener on your server:

```bash
NGROK_AUTHTOKEN=YOUR_TOKEN WEBHOOK_SECRET=YOUR_SECRET podswap -p YOUR_PROJECT_DIR -f PATH_TO_ACTION_FILE
```

When a push is done on main, the listener will run the commands specified on the action file.

Be aware of the implication here: `podswap` assumes the commands on your repository are safe and were written by you.

## Running with systemd

Replace the `<>` fields with your config and add it to `~/.config/systemd/user/podswap.service`:

```bash
[Unit]
Description=podswap
After=network-online.target

[Service]
ExecStart=/home/<YOUR USER>/go/bin/podswap -p PROJECT_DIR -f PROJECT_DIR/.github/workflows/podswap.yml
Environment="NGROK_AUTHTOKEN=<YOUR TOKEN>"
Environment="WEBHOOK_SECRET=<YOUR SECRET>"
Restart=on-failure
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
```

Start/enable the service:

```bash
systemctl --user --now enable podswap
```

To see the logs in real time:

```bash
journalctl --user -f -u podswap.service
```

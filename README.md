# podswap

PodsSwap is a lightweight application designed to update your containers with almost no downtime. By listening for the `push` event in your GitHub repository, PodsSwap re-builds and swaps your containers seamlessly.

This project is in its early stages, and we welcome feedback from users to help improve its functionality. 



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
  -build-cmd value
    	command to run after the webhook is triggered(default: "docker compose build").
  -deploy-cmd value
    	command to run after the build is finished(default: "docker compose up -d --force-recreate").
  -workdir value
    	working directory where containers will be deployed from(default: current directory).
```

Step-by-step setup:
- Set up an [ngrok account](https://dashboard.ngrok.com/signup) and get [your token](https://dashboard.ngrok.com/get-started/your-authtoken).
- Run the podswap command once with:
  - If you use docker: `NGROK_AUTHTOKEN=YOUR_TOKEN podswap`.
  - If you use podman: `NGROK_AUTHTOKEN=YOUR_TOKEN podswap -build-cmd "podman compose build" -deploy-cmd "podman compose up -d --force-recreate"`.
  - It's going to give you an URL in the terminal. Use it in the next step.

- Create a webhook by going to your repo's `Settings > Webhooks > Add webhook`.
  - In the `Payload URL` field type the URL you got in the previous step.
  - In the `Secret` field, you can make your own secret and type it there, you'll use it as an environment variable to run `podswap`.

- Run `podswap` changing the arguments and environment variables as needed:

```bash
NGROK_AUTHTOKEN=YOUR_TOKEN WEBHOOK_SECRET=YOUR_SECRET podswap -build-cmd "podman compose build" -deploy-cmd "podman compose up -d --force-recreate"
```

## Running with systemd

Replace the `<>` fields with your config and add it to `~/.config/systemd/user/podswap.service`:

```bash
[Unit]
Description=podswap
After=network-online.target

[Service]
ExecStart=/home/<YOUR USER>/go/bin/podswap -build-cmd "podman compose build" -deploy-cmd "podman compose up -d --force-recreate" --workdir /home/<YOUR USER>/<YOUR PROJECT FOLDER>
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

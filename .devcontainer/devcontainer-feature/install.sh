#!/bin/bash
set -e

echo "Activating feature 'devenv'"

LANGUAGE=${LANGUAGE:-undefined}
echo "The provided language is: $LANGUAGE"

export USER_HOME=/home/$_REMOTE_USER

# nvim config deps
dnf copr enable -y che/nerd-fonts && dnf install -y nerd-fonts
dnf install -y gcc cascadia-mono-nf-fonts git make npm nvim ripgrep shadow-utils unzip jetbrains-mono-fonts-all wl-clipboard

case $LANGUAGE in
    "c++")
        dnf install -y \
            glfw\
            glfw-devel\
            clang\
            clang-tools-extra
        ;;
    "go")
        dnf install -y golang
        useradd -m -s /bin/bash --create-home $_REMOTE_USER

        su - "$_REMOTE_USER" -c "
            go install github.com/air-verse/air@v1.62.0 &&
            go install github.com/go-delve/delve/cmd/dlv@v1.7.3 &&
            go install golang.org/x/tools/gopls@latest
            go install -v github.com/nicksnyder/go-i18n/v2/goi18n@latest
        "

        ;;
    *)
        echo "Invalid language"
        exit 1
        ;;
esac

echo PATH=$PATH:$USER_HOME/go/bin >> $USER_HOME/.bashrc
mkdir -p $USER_HOME/.config
git clone --branch mine https://github.com/douglascdev/kickstart.nvim $USER_HOME/.config/nvim
chown -R $_REMOTE_USER $USER_HOME

#!/bin/bash

if command -v etcd &> /dev/null
then
    echo "etcd is installed"
else
    ETCD_VER=v3.5.14
    echo "installing etcd"
    wget -q https://github.com/etcd-io/etcd/releases/download/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz &> /dev/null
    tar -xvf etcd-${ETCD_VER}-linux-amd64.tar.gz
    cd etcd-${ETCD_VER}-linux-amd64
    sudo mv etcd /usr/local/bin/
    sudo mv etcdctl /usr/local/bin/
    echo "etcd installed successfully"

    sudo tee /etc/systemd/system/etcd.service > /dev/null <<EOF
[Unit]
Description=etcd
Documentation=https://github.com/etcd-io/etcd
After=network-online.target

[Service]
User=root
Type=notify
ExecStart=/usr/local/bin/etcd

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable etcd
    sudo systemctl start etcd
fi
# Go
export PATH="$PATH:/usr/local/go/bin"
if  go version &> /dev/null
then
    echo "Go is installed"
    go version
else
    echo "Installing go"
    GO_VER=1.22.3
    wget --tries=0 --timeout=300 --waitretry=5 --read-timeout=20 https://mirrors.aliyun.com/golang/go${GO_VER}.linux-amd64.tar.gz  -O /tmp/go.tar.gz
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz

    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export GOPROXY=https://goproxy.cn,direct
    go version
    go mod tidy
fi

if command -v weave &> /dev/null
then
    echo "Weave is installed"
else
    echo "Installing weave"
    sudo wget -O /usr/local/bin/weave https://raw.githubusercontent.com/zettio/weave/master/weave && sudo chmod +x /usr/local/bin/weave

    sudo weave launch
    echo "Weave is installed successfully"
fi

if [ ! -x "$(command -v docker)" ]; then
  sudo apt-get update

  sudo apt-get install -y \
      apt-transport-https \
      ca-certificates \
      curl \
      gnupg-agent \
      software-properties-common

  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -

  sudo add-apt-repository \
     "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
     $(lsb_release -cs) \
     stable"

  sudo apt-get update

  sudo apt-get install -y docker-ce docker-ce-cli containerd.io
  sudo gpasswd -a "$USER" docker
  sudo getent passwd | while IFS=: read -r name _ uid gid _ home shell; do
    [ $uid -ge 1000 ] && sudo gpasswd -a "$name" docker
  done

  echo "Docker is installed."
else
  echo "Docker is installed."
fi

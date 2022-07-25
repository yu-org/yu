FROM ubuntu:20.04
LABEL maintainer="github.com/yu-org"

WORKDIR /workdir

ENV DEBIAN_FRONTEND=noninteractive

ENV TZ=Asia/Shanghai
ENV PATH /usr/local/go/bin:/usr/local/bin:/root/.local/bin:$PATH

RUN rm -fR /var/lib/apt/lists/* && \
    mkdir /var/lib/apt/lists/partial && \
    apt update && \
    apt install -y software-properties-common --fix-missing

# Install apt-fast
RUN add-apt-repository -y ppa:apt-fast/stable && apt update && apt install -y apt-fast

# Install gcc-11
RUN add-apt-repository -y ppa:ubuntu-toolchain-r/test && apt update && apt-fast install -y gcc-11 && \
    update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-11 10

# Install wget
RUN apt-fast install -y wget

# Install golang
RUN wget -O go.tar.gz https://go.dev/dl/go1.17.12.linux-amd64.tar.gz && tar zxvf go.tar.gz && rm -f go.tar.gz && mv go /usr/local
RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct && go env -w CGO_CFLAGS="-g -O2 -Wno-return-local-addr"

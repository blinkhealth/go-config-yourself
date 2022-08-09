FROM debian

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get -yq install curl
RUN last_release="https://api.github.com/repos/blinkhealth/go-config-yourself/releases/latest" && \
    version=$(curl --silent "$last_release" | awk -F'"' '/tag_name/{print $4}' ) && \
    curl -vOL https://github.com/blinkhealth/go-config-yourself/releases/download/$version/gcy-linux-amd64.deb && \
    apt install ./gcy-linux-amd64.deb

ENTRYPOINT ["gcy"]

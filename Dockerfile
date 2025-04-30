FROM golang:1.24-alpine3.21@sha256:7772cb5322baa875edd74705556d08f0eeca7b9c4b5367754ce3f2f00041ccee

WORKDIR /app

SHELL ["/bin/ash", "-o", "pipefail", "-c"]

ARG \
  # renovate: datasource=repology depName=alpine_3_21/ca-certificates
  CA_CERTIFICATES_VERSION="20241121-r1" \
  # renovate: datasource=repology depName=alpine_3_21/dpkg
  DPKG_VERSION="1.22.11-r0" \
  # renovate: datasource=repology depName=alpine_3_21/gnupg
  GNUPG_VERSION="2.4.7-r0" \
  # renovate: datasource=repology depName=alpine_3_21/curl
  CURL_VERSION="8.12.1-r1" \
  # renovate: datasource=repology depName=alpine_3_21/bash
  BASH_VERSION="5.2.37-r0" \
  # renovate: datasource=repology depName=alpine_3_21/openssl
  OPENSSL_VERSION="3.3.3-r0"

# renovate: datasource=github-releases depName=gosu packageName=tianon/gosu
ENV GOSU_VERSION=1.17
RUN apk add --no-cache --virtual .gosu-deps \
    ca-certificates=="${CA_CERTIFICATES_VERSION}" \
    dpkg=="${DPKG_VERSION}" \
    gnupg=="${GNUPG_VERSION}" && \
	dpkgArch="$(dpkg --print-architecture | awk -F- '{ print $NF }')" && \
	wget -q -O /usr/local/bin/gosu "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-$dpkgArch" && \
	wget -q -O /usr/local/bin/gosu.asc "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-$dpkgArch.asc" && \
	GNUPGHOME="$(mktemp -d)" && \
	export GNUPGHOME && \
	gpg --batch --keyserver hkps://keys.openpgp.org --recv-keys B42F6819007F00F88E364FD4036A9C25BF357DD4 && \
	gpg --batch --verify /usr/local/bin/gosu.asc /usr/local/bin/gosu && \
	gpgconf --kill all && \
	rm -rf "$GNUPGHOME" /usr/local/bin/gosu.asc && \
	apk del --no-network .gosu-deps && \
	chmod +x /usr/local/bin/gosu

WORKDIR /app

RUN adduser -S -G nobody isle-fcrepo-fs

RUN apk update && \
    apk add --no-cache \
      curl=="${CURL_VERSION}" \
      bash=="${BASH_VERSION}" \
      ca-certificates=="${CA_CERTIFICATES_VERSION}" \
      openssl=="${OPENSSL_VERSION}"

COPY . ./

RUN chown -R isle-fcrepo-fs:nobody /app

RUN go mod download && \
  go build -o /app/isle-fcrepo-fs && \
  go clean -cache -modcache && \
  ./ca-certs.sh

ENTRYPOINT ["/bin/bash"]
CMD ["/app/docker-entrypoint.sh"]

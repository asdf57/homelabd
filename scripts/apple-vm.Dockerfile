FROM golang:1.26-bookworm

# `container machine` boots /sbin/init. Application-oriented OCI images such
# as the standard Go image omit it, so add systemd and configure it for a VM.
ENV container=container

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        dbus \
        iproute2 \
        iputils-ping \
        pciutils \
        sudo \
        systemd \
        systemd-sysv \
        util-linux && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN > /etc/machine-id
RUN > /var/lib/dbus/machine-id

RUN systemctl set-default multi-user.target && \
    systemctl mask \
        console-getty.service \
        dev-hugepages.mount \
        sys-fs-fuse-connections.mount \
        systemd-tmpfiles-setup.service \
        systemd-update-utmp.service

FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/kubevirt-vm-operator \
    USER_UID=1001 \
    USER_NAME=kubevirt-vm-operator

# install operator binary
COPY build/_output/bin/kubevirt-vm-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}

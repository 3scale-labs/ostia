FROM registry.access.redhat.com/ubi8/ubi-minimal

ENV OPERATOR=/usr/local/bin/ostia-operator \
    USER_UID=1001 \
    USER_NAME=ostia-operator

# install operator binary
COPY build/_output/bin/ostia-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}

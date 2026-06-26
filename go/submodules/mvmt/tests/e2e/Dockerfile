# mvmt end-to-end test image. Built by tests/e2e/harness.ts after the host has
# run `npm run build && npm pack` and copied the tarball as mvmt.tgz into the
# build context.
FROM node:20-alpine

WORKDIR /opt/mvmt
COPY mvmt.tgz /tmp/mvmt.tgz
RUN apk add --no-cache curl
RUN npm install -g --silent /tmp/mvmt.tgz && rm /tmp/mvmt.tgz

# Tests bind-mount their per-scenario data into /workdir and use HOME=/workdir
# so config + session-token files stay isolated under that directory.
RUN mkdir -p /workdir
WORKDIR /workdir
ENV HOME=/workdir

CMD ["sleep", "infinity"]

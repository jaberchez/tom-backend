# sudo podman build --build-arg APP_VERSION=v1.3 -t quay.io/jberchez-redhat/tom-backend:v1.3 . && \
# sudo podman push quay.io/jberchez-redhat/tom-backend:v1.3

FROM golang:1.17.5 AS builder

RUN mkdir /tmp/tom-backend

COPY . /tmp/tom-backend/

WORKDIR /tmp/tom-backend

RUN CGO_ENABLED=0 GOOS=linux go build -o tom-backend main.go

FROM centos:8

ARG APP_VERSION=v1.0
ENV APP_VERSION=${APP_VERSION}

USER root

# Copy app from builder image
COPY --from=builder /tmp/tom-backend/tom-backend /usr/local/bin/

RUN chmod +x /usr/local/bin/tom-backend

RUN yum update -y && \
    yum install -y curl && \
    yum clean all
    
CMD ["/usr/local/bin/tom-backend"]

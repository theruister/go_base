FROM golang:1.27 AS build

ENV CGO_ENABLED=0

RUN mkdir -p /goBase
ADD . /goBase

WORKDIR /goBase/

#RUN go mod vendor
RUN go build /goBase/cmd/base

FROM alpine AS final
RUN mkdir /app
COPY --from=build /goBase /app/
#COPY --from=build /goBase/internal/scripts /app/
#COPY --from=build /goBase/config/config.yml /etc/go-base-app/

WORKDIR /app
CMD ["/app/base"]

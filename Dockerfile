FROM golang
WORKDIR learn/cmd/main/
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /learn/cmd/main/
EXPOSE 8080
CMD ["/learn/cmd/main/"]
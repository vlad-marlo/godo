FROM golang:1.20.1-alpine
WORKDIR /app

# download modules
COPY ../go.mod .
COPY ../go.sum .
RUN go mod download
RUN go mod verify

# migrate database
COPY ../migrations ./
RUN go install github.com/jackc/tern@latest
CMD ["tern", "migrate"]

# copy project files
COPY .. .

# build and run binarie
RUN go build --o server cmd/server/main.go
CMD ["./server", "-bind-port=8080"]
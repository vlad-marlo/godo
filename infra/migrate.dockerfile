FROM golang:1.20.1-alpine
WORKDIR /migrate
# migrate database
COPY ../migrations ./
RUN go install github.com/jackc/tern@latest
CMD ["tern", "migrate"]
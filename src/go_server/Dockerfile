FROM golang

WORKDIR /go_server

COPY go.mod go.sum ./

# Install deps
RUN go mod download

# Copy project files
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /vb-se

# App port
EXPOSE 3002

# Env vars
ENV PORT 3002

CMD [ "/vb-se"]

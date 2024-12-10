FROM golang

# Install python
RUN apt-get update
RUN apt-get -y install python3
RUN apt-get -y install python3-setuptools
RUN apt-get -y install python3-pip
RUN apt-get -y install python3-venv
RUN pip install youtube_transcript_api --break-system-packages
RUN pip install argparse --break-system-packages
RUN pip install yt-dlp --break-system-packages

WORKDIR /go_server

#RUN python3 -m venv vb-se-venv
#RUN vb-se-venv/bin/pip install argparse youtube_transcript_api yt-dlp

COPY go_server/go.mod go_server/go.sum ./

# Install deps
RUN go mod download

# Copy project files
COPY go_server .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build --tags fts5 -o /vb-se

# App port
EXPOSE 3002

# Env vars
ENV PORT 3002
ENV SQLITE_PATH /go_server/db/vb-se.db

# Make the database directory
RUN mkdir -p db

CMD [ "/vb-se"]

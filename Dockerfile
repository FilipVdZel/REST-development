
# Alpine is chosen because of its smaller size
# From Docker website
FROM golang:latest

# Creates working directory on the Docker image
WORKDIR /app

# Download necessary Go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy src files to working dir in Docker image
COPY *.go ./

# Build the application binary 
RUN go build -o /restserver

# Open port 8081 to be accesseble outside container
EXPOSE 8081

# This is the command that will execute when this image
# is used to start a container
CMD [ "/restserver"]


FROM golang:1.26.5-alpine

# Set destination
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

# Port 
EXPOSE 8080

# Run
CMD ["/worker"]


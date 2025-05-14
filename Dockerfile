######################################
### Stage 1: Build the application ###
######################################
FROM golang:1.22.8 AS builder

# Add Maintainer Info
LABEL maintainer="devops@runsystem.id"

# Set the Current Working Directory inside the container
WORKDIR /go/src/app

# Copy the source code
ADD . /go/src/app

# Ensure version.txt exists
RUN if [ ! -f version.txt ]; then touch version.txt; fi

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./main



####################################
### Stage 2: Run the application ###
####################################
FROM alpine:3 AS runner

# Add Maintainer Info
LABEL maintainer="devops@runsystem.id"

# Set timezone
RUN apk add --no-cache tzdata
ENV TZ=Asia/Jakarta
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy required files from builder
COPY --from=builder /go/src/app/config.yaml /app/config.yaml
COPY --from=builder /go/src/app/.env.sample /app/.env
COPY --from=builder /go/src/app/main /app/main
COPY --from=builder /go/src/app/version.txt /app/version.txt

EXPOSE 8000

# Start the application
CMD ["./main", "service"]

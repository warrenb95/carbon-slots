# Carbon Slots

A Go server that exposes an HTTP API for finding low-carbon electricity time slots.  
It integrates with the UK Carbon Intensity API and provides endpoints under `/api/v1/slots`.

## Requirements

- Go 1.24
- (Optional) Docker

## Running the Server (Go)

```sh
go run ./cmd/server
```

The server will start on port `3000` by default.

## Running with Docker

Build the Docker image:

```sh
docker build -t carbon-slots .
```

Run the container:

```sh
docker run -p 3000:3000 carbon-slots
```

## API

- `GET /api/v1/slots`  
  Returns available low-carbon slots.

  **Query Parameters:**

  - `duration` (int, optional): Duration in minutes for the slot. Default is `30`.
  - `contineous` (bool, optional): If `true`, ensures the slot duration is continuous. Default is `false`.

  **Example:**
  curl "<http://localhost:3000/api/v1/slots?duration=60&contineous=true>"

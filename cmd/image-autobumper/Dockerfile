FROM europe-docker.pkg.dev/kyma-project/prod/testimages/alpine-git:v20240924-6fb36f45

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o image-autobumper ./cmd/image-autobumper

ENTRYPOINT ["./image-autobumper"]

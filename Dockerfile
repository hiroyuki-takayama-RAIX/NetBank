# --- container to build ---
FROM golang:1.21.1-bullseye as deploy-builder

WORKDIR /app

COPY . ./
RUN go mod download
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags '-s -w -X main.version=1.0.0' main.go
# ldflag '-s' omits symbol table and debug info.
# ldflag '-w' omits symbol table in DWWARF.
# -trimpath remove file system in binary.
# CGO_ENABLED=0 disable cgo.
# ldflag '-X' add version into binary.


# --- container to deploy ---
FROM golang:1.21.1-bullseye as deploy
RUN apt update
ENV Env=prod
COPY --from=deploy-builder ./app/main ./app/main
WORKDIR /app
CMD ["go","run","main"]



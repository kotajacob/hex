FROM golang:1.21

WORKDIR /app
COPY . ./

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false .

EXPOSE 4000

ENTRYPOINT ["/app/hex"]
#CMD ["-hb","https://lemmygrad.ml/api/v3/"] #Optional

# Move this to the root of the repo and run something like this to use:
# sudo docker build --tag diethex .
# sudo docker run -d --publish 4000:4000 diethex

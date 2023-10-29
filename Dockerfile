FROM golang:latest

RUN apt-get update -qq

RUN apt-get install -y -qq libtesseract-dev libleptonica-dev

RUN apt-get install -y -qq \
  tesseract-ocr-eng \
  tesseract-ocr-deu \
  tesseract-ocr-jpn

COPY go.mod ${GOPATH}/
COPY go.sum ${GOPATH}/
COPY main.go ${GOPATH}/


WORKDIR ${GOPATH}/
ENTRYPOINT ["go", "run", "main.go"]

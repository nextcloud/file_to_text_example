FROM golang:latest

RUN apt-get update -qq

RUN apt-get install -y -qq libtesseract-dev libleptonica-dev

RUN apt-get install -y -qq \
  tesseract-ocr-eng \
  tesseract-ocr-deu \
  tesseract-ocr-jpn

COPY go.mod ${GOPATH}/app/
COPY go.sum ${GOPATH}/app/
COPY main.go ${GOPATH}/app/


WORKDIR ${GOPATH}/app/
ENTRYPOINT ["go", "run", "main.go"]

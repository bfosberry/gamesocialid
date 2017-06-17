FROM gobuffalo/buffalo:latest
RUN mkdir -p $GOPATH/src/github.com/bfosberry/gamesocialid
WORKDIR $GOPATH/src/github.com/bfosberry/gamesocialid
ADD package.json .
RUN npm install
ADD . .
RUN buffalo build -o bin/app
CMD ./bin/app

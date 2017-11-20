install: 
	go get github.com/Masterminds/glide
	go get github.com/golang/mock/mockgen
	glide install -v

dist/keyval:
	go build -o dist/keyval cmd/keyval/*.go

clean: 
	rm -f dist/keyval

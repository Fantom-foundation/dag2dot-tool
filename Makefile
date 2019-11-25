
# build
build :
	go build -o bin/dot-tool ./cmd/dot-tool 

#test
.PHONY : test
test :
	go test ./...

#clean
.PHONY : clean
clean :
	rm ./bin/dot-tool
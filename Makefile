install_dir = ~/.local/bin

deployment: 
	go build

.PHONY: install
install:
	mv deployment $(install_dir)

.PHONY: clean
clean:
	rm deployment

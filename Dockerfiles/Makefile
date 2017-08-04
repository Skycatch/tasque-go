ifeq ($(OS),Windows_NT)
	MAKEFILE = Makefile.win
else
	MAKEFILE = Makefile.linux
endif

build:
	make -f $(MAKEFILE) build

push:
	make -f $(MAKEFILE) push
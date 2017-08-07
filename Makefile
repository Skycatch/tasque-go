
ifeq ($(OS),Windows_NT)
	MAKEFILE = Makefile.win
else
	MAKEFILE = Makefile.linux
endif

build:
	make -f $(MAKEFILE) build

upload:
	make -f $(MAKEFILE) upload

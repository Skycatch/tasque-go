VERSIONS=4.8 6.10
VARIANTS=trusty

build:
	for version in $(VERSIONS); do \
		for variant in $(VARIANTS); do \
			TASQUE_VERSION=$(TASQUE_VERSION) \
			LANGUAGE=$(LANGUAGE) \
			VERSION=$$version \
			VARIANT=$$variant \
			make -C $$version/$$variant; \
		done \
	done

push:
	for version in $(VERSIONS); do \
		for variant in $(VARIANTS); do \
			TASQUE_VERSION=$(TASQUE_VERSION) \
			LANGUAGE=$(LANGUAGE) \
			VERSION=$$version \
			VARIANT=$$variant \
			make push -C $$version/$$variant; \
		done \
	done

build:
	cp ../../../../tasque ./
	docker build -t skycatch/tasque:$(VERSION)-$(TASQUE_VERSION)-$(VARIANT) \
							 -t skycatch/tasque:$(VERSION)-$(VARIANT) \
							 -t skycatch/tasque:$(VARIANT) \
							  .
	rm tasque

push:
	docker push skycatch/tasque:$(VERSION)-$(TASQUE_VERSION)-$(VARIANT)
	docker push skycatch/tasque:$(VARIANT)

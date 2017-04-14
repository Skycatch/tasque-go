build:
	cp ../../../../tasque ./
	docker build --force-rm --cpu-quota=1800000 --cpuset-cpus="0-3" --cpu-shares=102400 --memory=50GB \
							 -t skycatch/tasque:$(VERSION)-$(TASQUE_VERSION)-$(VARIANT) \
							 -t skycatch/tasque:$(VERSION)-$(VARIANT) \
							 -t skycatch/tasque:$(VARIANT) \
							  .
	rm tasque

push:
	docker push skycatch/tasque:$(VERSION)-$(TASQUE_VERSION)-$(VARIANT)
	docker push skycatch/tasque:$(VARIANT)

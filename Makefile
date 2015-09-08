.PHONY=clean release

OUTDIR=out
INSTALLDIR=$(OUTDIR)/opt/towncrier
CONTROLDIR=$(OUTDIR)/control

TOWNCRIER_TARGET_PATH=$(INSTALLDIR)/bin/towncrier
GOOSE_TARGET_PATH=$(INSTALLDIR)/bin/goose
REVISION_TARGET_PATH=$(INSTALLDIR)/REVISION
BUILD_TIME_TARGET_PATH=$(INSTALLDIR)/BUILD_TIME
CONTROL_TARGET_PATH=$(CONTROLDIR)/control

SHA=$(shell git rev-parse HEAD)
SHORT_SHA=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u +%Y%m%d%H%M)

release: $(TOWNCRIER_TARGET_PATH) $(GOOSE_TARGET_PATH) $(REVISION_TARGET_PATH) $(BUILD_TIME_TARGET_PATH) $(CONTROL_TARGET_PATH)
	mkdir -p $(INSTALLDIR)/config
	cp -r db $(INSTALLDIR)
	rm -f $(INSTALLDIR)/db/towncrier.development.db

	cp LICENSE $(INSTALLDIR)/towncrier-license
	echo "goose is distributed under the MIT license. See https://bitbucket.org/liamstask/goose for details." > $(INSTALLDIR)/goose-license

	cd $(CONTROLDIR) && tar --owner=0 --group=0 -czvf control.tar.gz control
	mv $(CONTROLDIR)/control.tar.gz $(OUTDIR)

	cd $(OUTDIR) && tar --owner=0 --group=0 -czvf data.tar.gz opt

	cd $(OUTDIR) && echo 2.0 > debian-binary

	cd $(OUTDIR) && ar r towncrier-$(SHORT_SHA).deb debian-binary control.tar.gz data.tar.gz

$(CONTROL_TARGET_PATH): clean
	mkdir -p $(CONTROLDIR)
	sed "s/0\.0\.1/0.0.1+$(BUILD_TIME)+$(SHORT_SHA)/g" debian/control > $(CONTROL_TARGET_PATH)

$(REVISION_TARGET_PATH): clean
	git diff --quiet HEAD ; if [ $$? -eq 0 ]; then echo $(SHA) > $(REVISION_TARGET_PATH); else echo "$(SHA)-dirty" > $(REVISION_TARGET_PATH); fi

$(BUILD_TIME_TARGET_PATH): clean
	echo $(BUILD_TIME) > $(BUILD_TIME_TARGET_PATH)

$(TOWNCRIER_TARGET_PATH): clean
	godep go build -o $(TOWNCRIER_TARGET_PATH) gitlab.com/shuhao/towncrier/application

$(GOOSE_TARGET_PATH): clean
	go get -v bitbucket.org/liamstask/goose/cmd/goose
	cp $(GOPATH)/bin/goose $(GOOSE_TARGET_PATH)

clean:
	rm -rf $(OUTDIR)

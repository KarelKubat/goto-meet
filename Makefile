BINDIR=/usr/local/bin
foo:
	@echo 'Make what? Choose:'
	@echo '  make local   - creates local binary'
	@echo '  make install - and installs to ${BINDIR}'
	@echo '  make reload  - and kills running copy, so that a new is reloaded'

local:
	go build goto-meet

install: local ${BINDIR}/goto-meet
	sudo install goto-meet ${BINDIR}/

reload: install
	killall goto-meet

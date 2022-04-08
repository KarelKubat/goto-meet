BINDIR=/usr/local/bin
foo:
	@echo 'Make what? Choose:'
	@echo '  make local   - creates local binary'
	@echo '  make install - and installs to ${BINDIR}'
	@echo '  make reload  - and kills running copy, so that a new is reloaded'

local:
	go build goto-meet.go

goto-meet: local

install: ${BINDIR}/goto-meet

${BINDIR}/goto-meet: goto-meet
	sudo install goto-meet ${BINDIR}/goto-meet

# Reload will only work if you have some mechanism of detecting that a previous
# run of goto-meet stopped and of starting a new one, like MacOSX's launchctl.
reload: install
	killall goto-meet

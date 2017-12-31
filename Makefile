default: build clean

clean:
	rm checker

build:
	go build -o checker -race

setup:
	mkdir -p ~/primenotify
	cp checker ~/primenotify/
	cp config.json ~/primenotify/

launchctl:
	./generate_launchfile
	launchctl load /Library/LaunchAgents/com.github.primenotify.plist

install: build setup launchctl clean

uninstall:
	rm -rf ~/primenotify
	launchctl unload /Library/LaunchAgents/com.github.primenotify.plist
	sudo rm -f /Library/LaunchAgents/com.github.primenotify.plist


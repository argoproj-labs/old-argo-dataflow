tag=$(shell git describe --tags --abbrev=0)

install: env
	./venv/bin/pip3 install --use-feature=in-tree-build .
	./venv/bin/pip3 install -r ./requirements.txt

env:
	python3 -m venv venv

test:
	./venv/bin/python3 -m pytest

build:
	sed -i "" "s/version='.*'/version='$(tag)'/" setup.py

publish: build
	rm -Rf dist
	python setup.py sdist
	twine upload dist/*
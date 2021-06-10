tag=$(shell git describe --tags --abbrev=0)

build:
	sed -i "" "s/version='.*'/version='$(tag)'/" setup.py

publish: build
	rm -Rf dist
	python setup.py sdist
	twine upload dist/*
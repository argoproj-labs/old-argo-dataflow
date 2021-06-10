tag=$(shell git describe --tags --abbrev=0)

publish:
	sed -i "" "s/version='.*'/version='$(tag)'/" setup.py
	python setup.py sdist
	twine upload dist/*
install: env
	./venv/bin/pip3 install --use-feature=in-tree-build .
	./venv/bin/pip3 install -r ./requirements.txt

env:
	python3 -m venv venv

test:
	./venv/bin/python3 -m pytest

publish: build
	rm -Rf dist
	python3 -m build
	python3 -m twine upload dist/*
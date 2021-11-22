# Adding custom image

Dataflow allows custom images to be used instead of the default runtimes. For this, the
required images must be built, using some standard Dockerfile components.

## Image dockerfile
As an example, this dockerfile is used to create a Python image.
```
# import base image
FROM <image_name>
USER root

# install dumb-init
RUN wget -O ./dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.5/dumb-init_1.2.5_x86_64
RUN chmod +x /dumb-init

# create directories
RUN mkdir /.cache /.local
ADD . /workspace

# install any custom requirements
COPY requirements.txt .
RUN python -m pip install --upgrade pip && pip install -r requirements.txt

# install Python dsl dependencies
ADD dsls/python /workspace/.dsl
RUN cd /workspace/.dsl && python -m pip install --use-feature=in-tree-build .

# install Python sdk dependencies
ADD sdks/python /workspace/.sdk
RUN cd /workspace/.sdk && python -m pip install --use-feature=in-tree-build .

# Provide necessary permissions
RUN chown -R 9653 /.cache /.local /workspace
WORKDIR /workspace

USER 9653:9653
ENV PYTHONUNBUFFERED 1

# Provide entrypoint
RUN chmod +x /workspace/entrypoint.sh
ENTRYPOINT ["/dumb-init", "--"]
CMD ["/workspace/entrypoint.sh"]
```

## Dummy Handler file
A dummy handler file needs to be referenced which will actually be overridden by a 
user defined handler in the Python dsl file.
```
def handler(msg, context):
    return ("This is a dummy handler:" + msg.decode("UTF-8")).encode("UTF-8")
```

## Main runner file
A main runner file which acts as a starting process for the container.
```
from argo_dataflow_sdk import ProcessHandler

# Referencing the dummy handler created above
from handler import handler

if __name__ == '__main__':
    processHandler = ProcessHandler()
    processHandler.start(handler)
```

## Entrypoint shell file
```
#!/bin/sh
set -eux

# override the dummy handler
cp /var/run/argo-dataflow/handler handler.py

# Start the process handler
python main.py
```
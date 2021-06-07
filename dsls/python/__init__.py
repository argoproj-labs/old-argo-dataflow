import inspect
import sys
import yaml


def str_presenter(dumper, data):
    if len(data.splitlines()) > 1 or '"' in data or "'" in data:
        return dumper.represent_scalar('tag:yaml.org,2002:str', data, style='|')
    return dumper.represent_scalar('tag:yaml.org,2002:str', data)


yaml.add_representer(str, str_presenter)


class PipelineBuilder():
    def __init__(self, name):
        self._name = name
        self._annotations = {}
        self._steps = []

    def annotate(self, name, value):
        self._annotations[name] = value
        return self

    def describe(self, value):
        return self.annotate('dataflow.argoproj.io/description', value)

    def step(self, step):
        self._steps.append(step)
        return self

    def build(self):
        return {
            'apiVersion': 'dataflow.argoproj.io/v1alpha1',
            'kind': 'Pipeline',
            'metadata': {
                'name': self._name,
                'annotations': self._annotations
            },
            'spec': {
                'steps': [x.build() for x in self._steps]
            }
        }

    def dump(self):
        sys.stdout.write((yaml.dump(self.build())))


def pipeline(name):
    return PipelineBuilder(name)


class LogSink:
    def build(self):
        return {'log': {}}


class HTTPSink:
    def __init__(self, url):
        self._url = url

    def build(self):
        return {'http': {'url': self._url}}


class KafkaSink:
    def __init__(self, subject):
        self._subject = subject

    def build(self):
        return {'kafka': {'topic': self._subject}}


class STANSink:
    def __init__(self, topic):
        self._topic = topic

    def build(self):
        return {'stan': {'subject': self._topic}}


class Step:
    def __init__(self, name, sources=[], volumes=[]):
        self._name = name
        self._sources = sources
        self._sinks = []
        self._scale = {}
        self._volumes = volumes
        self._terminator = False

    def log(self):
        self._sinks.append(LogSink())
        return self

    def http(self, url):
        self._sinks.append(HTTPSink(url))
        return self

    def kafka(self, subject):
        self._sinks.append(KafkaSink(subject))
        return self

    def scale(self, minReplicas, maxReplicas, replicaRatio):
        self._scale = {
            'minReplicas': minReplicas,
            'maxReplicas': maxReplicas,
            'replicaRatio': replicaRatio
        }
        return self

    def stan(self, topic):
        self._sinks.append(STANSink(topic))
        return self

    def terminator(self):
        self._terminator = True
        return self

    def build(self):
        y = {
            'name': self._name,
        }
        if len(self._sources):
            y['sources'] = [x.build() for x in self._sources]
        if len(self._sinks):
            y['sinks'] = [x.build() for x in self._sinks]
        if len(self._scale) > 0:
            y['scale'] = self._scale
        if len(self._volumes) > 0:
            y['volumes'] = self._volumes
        if self._terminator:
            y['terminator'] = True
        return y


class CatStep(Step):
    def __init__(self, name, sources):
        super().__init__(name, sources=sources)

    def build(self):
        x = super().build()
        x['cat'] = {}
        return x


class ContainerStep(Step):
    def __init__(self, name, image, args, fifo=False, volumes=[], volumeMounts=[], sources=[]):
        super().__init__(name, sources=sources, volumes=volumes)
        self._image = image
        self._args = args
        self._fifo = fifo
        self._volumeMounts = volumeMounts

    def build(self):
        x = super().build()
        c = {
            'image': self._image,
            'args': self._args
        }
        if self._fifo:
            c['in'] = {'fifo': True}
        if len(self._volumeMounts) > 0:
            c['volumeMounts'] = self._volumeMounts
        x['container'] = c
        return x


class ExpandStep(Step):
    def __init__(self, name, sources=[]):
        super().__init__(name, sources=sources)

    def build(self):
        x = super().build()
        x['expand'] = {}
        return x


class FilterStep(Step):
    def __init__(self, name, filter, sources=[]):
        super().__init__(name, sources=sources)
        self._filter = filter

    def build(self):
        x = super().build()
        x['filter'] = self._filter
        return x


class GitStep(Step):
    def __init__(self, name, url, branch, path, image, sources=[]):
        super().__init__(name, sources=sources)
        self._url = url
        self._branch = branch
        self._path = path
        self._image = image

    def build(self):
        x = super().build()
        x['git'] = {
            'url': self._url,
            'branch': self._branch,
            'path': self._path,
            'image': self._image
        }
        return x


class GroupStep(Step):
    def __init__(self, name, key, format, endOfGroup, storage, sources=[]):
        super().__init__(name, sources=sources, volumes=[{'emptyDir': {}, 'name': 'group'}])
        self._key = key
        self._format = format
        self._endOfGroup = endOfGroup
        self._storage = storage

    def build(self):
        x = super().build()
        x['group'] = {
            'key': self._key,
            'format': self._format,
            'endOfGroup': self._endOfGroup,
            'storage': {
                'name': 'groups'
            },
        }
        return x


class FlattenStep(Step):
    def __init__(self, name, sources=[]):
        super().__init__(name, sources=sources)

    def build(self):
        x = super().build()
        x['flatten'] = {}
        return x


class HandlerStep(Step):
    def __init__(self, name, handler=None, code=None, runtime=None, sources=[]):
        super().__init__(name, sources=sources)
        if handler:
            self._code = inspect.getsource(handler)
        else:
            self._code = code
        if runtime:
            self._runtime = runtime
        else:
            self._runtime = 'python3-9'

    def build(self):
        x = super().build()
        x['handler'] = {
            'runtime': self._runtime,
            'code': self._code,
        }
        return x


class MapStep(Step):
    def __init__(self, name, map, sources=[]):
        super().__init__(name, sources=sources)
        self._map = map

    def build(self):
        x = super().build()
        x['map'] = self._map
        return x


class Source:

    def cat(self, name):
        return CatStep(name, sources=[self])

    def container(self, name, image, args, fifo=False, volumes=[], volumeMounts=[]):
        return ContainerStep(name, sources=[self], image=image, args=args, fifo=fifo, volumes=volumes,
                             volumeMounts=volumeMounts)

    def expand(self, name):
        return ExpandStep(name, sources=[self])

    def filter(self, name, filter):
        return FilterStep(name, filter, sources=[self])

    def git(self, name, url, branch, path, image):
        return GitStep(name, url, branch, path, image, sources=[self])

    def group(self, name, key, format, endOfGroup, storage):
        return GroupStep(name, key, format, endOfGroup, storage, sources=[self])

    def flatten(self, name):
        return FlattenStep(name, sources=[self])

    def handler(self, name, handler=None, code=None, runtime=None):
        return HandlerStep(name, handler=handler, code=code, runtime=runtime, sources=[self])

    def map(self, name, map):
        return MapStep(name, map, sources=[self])


def cat(name):
    return CatStep(name, [])


def container(name, image, args, fifo=False, volumes=[], volumeMounts=[]):
    return ContainerStep(name, sources=[], image=image, args=args, fifo=fifo, volumes=volumes,
                         volumeMounts=volumeMounts)


def expand(name):
    return ExpandStep(name)


def filter(name, filter):
    return FilterStep(name, filter)


def git(name, url, branch, path, image):
    return GitStep(name, url, branch, path, image)


def group(name, key, format, endOfGroup, storage):
    return GroupStep(name, key, format, endOfGroup, storage)


def flatten(name):
    return FlattenStep(name)


def handler(name, handler=None, code=None, runtime=None):
    return HandlerStep(name, handler, code, runtime)


def map(name, map):
    return MapStep(name, map)


class CronSource(Source):
    def __init__(self, schedule, layout):
        super().__init__()
        self._schedule = schedule
        self._layout = layout

    def build(self):
        x = {'schedule': self._schedule}
        if self._layout:
            x['layout'] = self._layout
        return {'cron': x}


class HTTPSource(Source):
    def __init__(self):
        super().__init__()

    def build(self):
        return {'http': {}}


class KafkaSource(Source):
    def __init__(self, topic, parallel=None):
        super().__init__()
        self._topic = topic
        self._parallel = parallel

    def build(self):
        y = {'topic': self._topic}
        if self._parallel:
            y['parallel'] = self._parallel
        return {'kafka': y}


class STANSource(Source):
    def __init__(self, subject):
        super().__init__()
        self._subject = subject

    def build(self):
        return {'stan': {'subject': self._subject}}


def cron(schedule, layout=''):
    return CronSource(schedule, layout)


def http():
    return HTTPSource()


def kafka(topic, parallel=None):
    return KafkaSource(topic, parallel)


def stan(subject):
    return STANSource(subject)

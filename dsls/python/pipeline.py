import getpass
import inspect
import json
import kubernetes
import yaml

RESOURCE_PLURAL = 'pipelines'
VERSION = 'v1alpha1'
GROUP = 'dataflow.argoproj.io'

DEFAULT_RUNTIME = 'python3-9'
GROUPS_VOLUME_NAME = 'groups'
USER = getpass.getuser()


def str_presenter(dumper, data):
    if '\n' in data or '"' in data or "'" in data:
        return dumper.represent_scalar('tag:yaml.org,2002:str', data, style='|')
    return dumper.represent_scalar('tag:yaml.org,2002:str', data)


yaml.add_representer(str, str_presenter)


class PipelineBuilder:
    def __init__(self, name):
        self._name = name
        self._resourceVersion = None
        self._namespace = None
        self._annotations = {}
        self._steps = []
        self.owner(USER)

    def annotate(self, name, value):
        self._annotations[name] = value
        return self

    def owner(self, value):
        return self.annotate('dataflow.argoproj.io/owner', value)

    def describe(self, value):
        return self.annotate('dataflow.argoproj.io/description', value)

    def namespace(self, namespace):
        self._namespace = namespace
        return self

    def step(self, step):
        self._steps.append(step)
        return self

    def dump(self):
        m = {
            'name': self._name,
            'annotations': self._annotations
        }
        if self._namespace:
            m['namespace'] = self._namespace
        if self._resourceVersion:
            m['resourceVersion'] = self._resourceVersion
        return {
            'apiVersion': 'dataflow.argoproj.io/v1alpha1',
            'kind': 'Pipeline',
            'metadata': m,
            'spec': {
                'steps': [x.dump() for x in self._steps]
            }
        }

    def yaml(self):
        return yaml.dump(self.dump())

    def json(self):
        return json.dumps(self.dump())

    def save(self):
        with open(self._name + '-pipeline.yaml', "w") as f:
            f.write(self.yaml())

    def start(self):
        # https://github.com/kubernetes-client/python
        kubernetes.config.load_kube_config()

        # https://github.com/kubernetes-client/python/blob/master/kubernetes/docs/CustomObjectsApi.md
        api = kubernetes.client.CustomObjectsApi()
        try:
            x = api.get_namespaced_custom_object(GROUP, VERSION, self._namespace, RESOURCE_PLURAL, self._name)
            self._resourceVersion = x['metadata']['resourceVersion']
            api.replace_namespaced_custom_object(GROUP, VERSION, self._namespace, RESOURCE_PLURAL, self._name,
                                                 self.dump())
            print('updated pipeline ' + self._namespace + '/' + self._name)
        except kubernetes.client.rest.ApiException as e:
            if e.status == 404:
                pass
                api.create_namespaced_custom_object(GROUP, VERSION, self._namespace, RESOURCE_PLURAL, self.dump())
                print('created pipeline ' + self._namespace + '/' + self._name)
        return self

    def watch(self):
        api = kubernetes.client.CustomObjectsApi()

        for event in kubernetes.watch.Watch().stream(api.list_namespaced_custom_object, GROUP, VERSION, self._namespace,
                                                     RESOURCE_PLURAL,
                                                     field_selector='metadata.name=' + self._name, watch=True):
            status = (event['object'].get('status') or {})
            print((status.get('phase') or 'Unknown') + ': ' + (status.get('message') or ''))

    def run(self):
        self.start()
        self.watch()

def pipeline(name):
    return PipelineBuilder(name)


class Sink:
    def __init__(self, name=None):
        self._name = name

    def dump(self):
        x = {}
        if self._name:
            x['name'] = self._name
        return x


class LogSink(Sink):
    def __init__(self, name=None):
        super().__init__(name)

    def dump(self):
        x = super().dump()
        x['log'] = {}
        return x


class HTTPSink(Sink):
    def __init__(self, url, name=None):
        super().__init__(name)
        self._url = url

    def dump(self):
        x = super().dump()
        x['http'] = {'url': self._url}
        return x


class KafkaSink(Sink):
    def __init__(self, subject, name=None):
        super().__init__(name)
        self._subject = subject

    def dump(self):
        x = super().dump()
        x['kafka'] = {'topic': self._subject}
        return x


class STANSink(Sink):
    def __init__(self, topic, name=None):
        super().__init__(name)
        self._topic = topic

    def dump(self):
        x = super().dump()
        x['stan'] = {'subject': self._topic}
        return x


class Step:
    def __init__(self, name, sources=[], volumes=[]):
        self._name = name
        self._sources = sources
        self._sinks = []
        self._scale = {}
        self._volumes = volumes
        self._terminator = False
        self._annotations = []

    def log(self, name=None):
        self._sinks.append(LogSink(name=name))
        return self

    def http(self, url, name=None):
        self._sinks.append(HTTPSink(url, name=name))
        return self

    def kafka(self, subject, name=None):
        self._sinks.append(KafkaSink(subject, name=name))
        return self

    def scale(self, minReplicas, maxReplicas, replicaRatio):
        self._scale = {
            'minReplicas': minReplicas,
            'maxReplicas': maxReplicas,
            'replicaRatio': replicaRatio
        }
        return self

    def stan(self, topic, name=None):
        self._sinks.append(STANSink(topic, name=name))
        return self

    def terminator(self):
        self._terminator = True
        return self

    def annotations(self, annotations):
        self._annotations = annotations
        return self

    def dump(self):
        y = {
            'name': self._name,
        }
        if len(self._sources):
            y['sources'] = [x.dump() for x in self._sources]
        if len(self._sinks):
            y['sinks'] = [x.dump() for x in self._sinks]
        if len(self._scale) > 0:
            y['scale'] = self._scale
        if len(self._volumes) > 0:
            y['volumes'] = self._volumes
        if self._terminator:
            y['terminator'] = True
        if self._annotations:
            # TODO - labels too please
            y['metadata'] = {
                'annotations': self._annotations
            }
        return y


class CatStep(Step):
    def __init__(self, name, sources):
        super().__init__(name, sources=sources)

    def dump(self):
        x = super().dump()
        x['cat'] = {}
        return x


class ContainerStep(Step):
    def __init__(self, name, image, args, fifo=False, volumes=[], volumeMounts=[], sources=[], env={}):
        super().__init__(name, sources=sources, volumes=volumes)
        self._image = image
        self._args = args
        self._fifo = fifo
        self._volumeMounts = volumeMounts
        self._env = env

    def dump(self):
        x = super().dump()
        c = {
            'image': self._image,
        }
        if len(self._args) > 0:
            c['args'] = self._args
        if self._fifo:
            c['in'] = {'fifo': True}
        if len(self._volumeMounts) > 0:
            c['volumeMounts'] = self._volumeMounts
        if self._env:
            c['env'] = [{'name': x, 'value': self._env[x]} for k, x in enumerate(self._env)]
        x['container'] = c
        return x


class ExpandStep(Step):
    def __init__(self, name, sources=[]):
        super().__init__(name, sources=sources)

    def dump(self):
        x = super().dump()
        x['expand'] = {}
        return x


class FilterStep(Step):
    def __init__(self, name, filter, sources=[]):
        super().__init__(name, sources=sources)
        self._filter = filter

    def dump(self):
        x = super().dump()
        x['filter'] = self._filter
        return x


class GitStep(Step):
    def __init__(self, name, url, branch, path, image, sources=[]):
        super().__init__(name, sources=sources)
        self._url = url
        self._branch = branch
        self._path = path
        self._image = image

    def dump(self):
        x = super().dump()
        x['git'] = {
            'url': self._url,
            'branch': self._branch,
            'path': self._path,
            'image': self._image
        }
        return x


def storageVolumes(storage=None):
    if storage:
        storage['name'] = GROUPS_VOLUME_NAME
        return [storage]
    return []


class GroupStep(Step):
    def __init__(self, name, key, format, endOfGroup, storage=None, sources=[]):
        super().__init__(name, sources=sources, volumes=storageVolumes(storage))
        self._key = key
        self._format = format
        self._endOfGroup = endOfGroup
        self._storage = storage

    def dump(self):
        x = super().dump()
        y = {
            'key': self._key,
            'format': self._format,
            'endOfGroup': self._endOfGroup,
        }
        if self._storage:
            y['storage'] = {
                'name': GROUPS_VOLUME_NAME
            }
        x['group'] = y
        return x


class FlattenStep(Step):
    def __init__(self, name, sources=[]):
        super().__init__(name, sources=sources)

    def dump(self):
        x = super().dump()
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
            self._runtime = DEFAULT_RUNTIME

    def dump(self):
        x = super().dump()
        x['handler'] = {
            'runtime': self._runtime,
            'code': self._code,
        }
        return x


class MapStep(Step):
    def __init__(self, name, map, sources=[]):
        super().__init__(name, sources=sources)
        self._map = map

    def dump(self):
        x = super().dump()
        x['map'] = self._map
        return x


class Source:
    def __init__(self, name=None, retryPolicy=None):
        self._name = name
        self._retryPolicy = retryPolicy

    def dump(self):
        x = {}
        if self._name:
            x['name'] = self._name
        if self._retryPolicy:
            x['retryPolicy'] = self._retryPolicy
        return x

    def cat(self, name):
        return CatStep(name, sources=[self])

    def container(self, name, image, args=[], fifo=False, volumes=[], volumeMounts=[], env={}):
        return ContainerStep(name, sources=[self], image=image, args=args, fifo=fifo, volumes=volumes,
                             volumeMounts=volumeMounts, env=env)

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


def container(name, image, args, fifo=False, volumes=[], volumeMounts=[], env={}):
    return ContainerStep(name, sources=[], image=image, args=args, fifo=fifo, volumes=volumes,
                         volumeMounts=volumeMounts, env=env)


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
    def __init__(self, schedule, layout, name=None, retryPolicy=None):
        super().__init__(name=name, retryPolicy=retryPolicy)
        self._schedule = schedule
        self._layout = layout

    def dump(self):
        x = super().dump()
        y = {'schedule': self._schedule}
        if self._layout:
            y['layout'] = self._layout
        x['cron'] = y
        return x


class HTTPSource(Source):
    def __init__(self, name=None, retryPolicy=None):
        super().__init__(name=name, retryPolicy=retryPolicy)

    def dump(self):
        x = super().dump()
        x['http'] = {}
        return x


class KafkaSource(Source):
    def __init__(self, topic, name=None, retryPolicy=None):
        super().__init__(name=name, retryPolicy=retryPolicy)
        self._topic = topic

    def dump(self):
        x = super().dump()
        x['topic'] = self._topic
        return {'kafka': x}


class STANSource(Source):
    def __init__(self, subject, name=None, retryPolicy=None):
        super().__init__(name=name, retryPolicy=retryPolicy)
        self._subject = subject

    def dump(self):
        x = super().dump()
        y = {'subject': self._subject}
        x['stan'] = y
        return x


def cron(schedule, layout=None, name=None, retryPolicy=None):
    return CronSource(schedule, layout=layout, name=name, retryPolicy=retryPolicy)


def http(name=None, retryPolicy=None):
    return HTTPSource(name=name, retryPolicy=retryPolicy)


def kafka(topic, name=None, retryPolicy=None):
    return KafkaSource(topic, name=name, retryPolicy=retryPolicy)


def stan(subject, name=None, retryPolicy=None):
    return STANSource(subject, name=name, retryPolicy=retryPolicy)

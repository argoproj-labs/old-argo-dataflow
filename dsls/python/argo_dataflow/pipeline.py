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
            x = api.get_namespaced_custom_object(
                GROUP, VERSION, self._namespace, RESOURCE_PLURAL, self._name)
            self._resourceVersion = x['metadata']['resourceVersion']
            api.replace_namespaced_custom_object(GROUP, VERSION, self._namespace, RESOURCE_PLURAL, self._name,
                                                 self.dump())
            print('updated pipeline ' + self._namespace + '/' + self._name)
        except kubernetes.client.rest.ApiException as e:
            if e.status == 404:
                pass
                api.create_namespaced_custom_object(
                    GROUP, VERSION, self._namespace, RESOURCE_PLURAL, self.dump())
                print('created pipeline ' + self._namespace + '/' + self._name)
        return self

    def watch(self):
        api = kubernetes.client.CustomObjectsApi()

        for event in kubernetes.watch.Watch().stream(api.list_namespaced_custom_object, GROUP, VERSION, self._namespace,
                                                     RESOURCE_PLURAL,
                                                     field_selector='metadata.name=' + self._name, watch=True):
            status = (event['object'].get('status') or {})
            phase = status.get('phase') or 'Unknown'
            message = status.get('message') or ''
            print(phase + ': ' + message)

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
    def __init__(self, url, name=None, insecureSkipVerify=None, headers=None):
        super().__init__(name)
        self._insecureSkipVerify = insecureSkipVerify
        self._url = url
        self._headers = headers

    def dump(self):
        x = super().dump()
        h = {'url': self._url}
        if self._headers:
            h['headers'] = self._headers
        if self._insecureSkipVerify:
            h['insecureSkipVerify'] = self._insecureSkipVerify
        x['http'] = h
        return x


class KafkaSink(Sink):
    def __init__(self, subject, name=None, a_sync=False, batchSize=None, linger=None, compressionType=None, acks=None):
        super().__init__(name)
        self._subject = subject
        self._a_sync = a_sync
        self._batchSize = batchSize
        self._linger = linger
        self._compressionType = compressionType
        self._acks = acks

    def dump(self):
        x = super().dump()
        y = {'topic': self._subject}
        if self._a_sync:
            y['async'] = True
        if self._batchSize:
            y['batchSize'] = self._batchSize
        if self._linger:
            y['linger'] = self._linger
        if self._compressionType:
            y['compressionType'] = self._compressionType
        if self._acks:
            y['acks'] = self._acks
        x['kafka'] = y
        return x


class STANSink(Sink):
    def __init__(self, topic, name=None):
        super().__init__(name)
        self._topic = topic

    def dump(self):
        x = super().dump()
        x['stan'] = {'subject': self._topic}
        return x


class JetStreamSink(Sink):
    def __init__(self, subject, name=None):
        super().__init__(name=name)
        self._subject = subject

    def dump(self):
        x = super().dump()
        x['jetstream'] = {'subject': self._subject}
        return x


class Step:
    def __init__(self, name, sources=None, sinks=None, volumes=None, terminator=False, sidecarResource=None):
        self._name = name or 'main'
        self._sources = sources or []
        self._sinks = sinks or []
        self._scale = None
        self._volumes = volumes or []
        self._terminator = terminator
        self._annotations = []
        self._sidecarResources = sidecarResource

    def log(self, name=None):
        self._sinks.append(LogSink(name=name))
        return self

    def http(self, url, name=None, insecureSkipVerify=None, headers=None):
        self._sinks.append(HTTPSink(
            url, name=name, insecureSkipVerify=insecureSkipVerify, headers=headers))
        return self

    def kafka(self, subject, name=None, a_sync=False, batchSize=None, linger=None, compressionType=None, acks=None):
        self._sinks.append(KafkaSink(subject, name=name, a_sync=a_sync, batchSize=batchSize, linger=linger,
                                     compressionType=compressionType, acks=acks))
        return self

    def scale(self, desiredReplicas, scalingDelay=None, peekDelay=None):
        self._scale = {
            'desiredReplicas': desiredReplicas,
        }
        if peekDelay:
            self._scale['peekDelay'] = peekDelay
        if scalingDelay:
            self._scale['scalingDelay'] = scalingDelay
        return self

    def stan(self, topic, name=None):
        self._sinks.append(STANSink(topic, name=name))
        return self

    def jetstream(self, subject, name=None):
        self._sinks.append(JetStreamSink(subject, name=name))
        return self

    def terminator(self):
        self._terminator = True
        return self

    def annotations(self, annotations):
        self._annotations = annotations
        return self

    def sidecarResources(self, sidecarResources):
        self._sidecarResources = sidecarResources
        return self

    def dump(self):
        y = {
            'name': self._name,
        }
        if len(self._sources):
            y['sources'] = [x.dump() for x in self._sources]
        if len(self._sinks):
            y['sinks'] = [x.dump() for x in self._sinks]
        if self._scale:
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
        if self._sidecarResources:
            y['sidecar'] = {
                'resources': self._sidecarResources
            }
        return y


class CatStep(Step):
    def __init__(self, name=None, sources=None, sinks=None):
        super().__init__(name=name, sources=sources, sinks=sinks)

    def dump(self):
        x = super().dump()
        x['cat'] = {}
        return x


class ContainerStep(Step):
    def __init__(self, name=None, image=None, args=None, fifo=False, volumes=None, volumeMounts=None, sources=None,
                 sinks=None,
                 env=None, resources=None,
                 terminator=False):
        super().__init__(name, sources=sources, sinks=sinks,
                         volumes=volumes, terminator=terminator)
        assert image
        self._image = image
        self._args = args or []
        self._fifo = fifo
        self._volumeMounts = volumeMounts or []
        self._env = env
        self._resources = resources

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
            c['env'] = [{'name': x, 'value': self._env[x]}
                        for k, x in enumerate(self._env)]
        if self._resources:
            c['resources'] = self._resources
        x['container'] = c
        return x


class DedupeStep(Step):
    def __init__(self, name=None, sources=None, sinks=None):
        super().__init__(name, sources=sources, sinks=sinks)

    def dump(self):
        x = super().dump()
        x['dedupe'] = {}
        return x


class ExpandStep(Step):
    def __init__(self, name=None, sources=None, sinks=None):
        super().__init__(name, sources=sources, sinks=sinks)

    def dump(self):
        x = super().dump()
        x['expand'] = {}
        return x


class FilterStep(Step):
    def __init__(self, name=None, expression=None, sources=None, sinks=None):
        super().__init__(name, sources=sources, sinks=sinks)
        assert expression
        self._expression = expression

    def dump(self):
        x = super().dump()
        x['filter'] = {
            'expression': self._expression
        }
        return x


class GitStep(Step):
    def __init__(self, name=None, url=None, branch=None, path=None, image=None, sources=None, sinks=None, env=None,
                 terminator=False,
                 command=None):
        super().__init__(name, sources=sources, sinks=sinks, terminator=terminator)
        assert url
        assert image
        self._url = url
        self._branch = branch or 'main'
        self._path = path or '.'
        self._image = image
        self._env = env
        self._command = command

    def dump(self):
        x = super().dump()
        y = {
            'url': self._url,
            'branch': self._branch,
            'path': self._path,
            'image': self._image
        }
        if self._command:
            y['command'] = self._command
        if self._env:
            y['env'] = self._env
        x['git'] = y
        return x


def storageVolumes(storage=None):
    if storage:
        storage['name'] = GROUPS_VOLUME_NAME
        return [storage]
    return []


class GroupStep(Step):
    def __init__(self, name=None, key=None, format=None, endOfGroup=None, storage=None, sources=None, sinks=None):
        super().__init__(name, sources=sources, sinks=sinks, volumes=storageVolumes(storage))
        assert key
        assert format
        assert endOfGroup
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
    def __init__(self, name=None, sources=None, sinks=None):
        super().__init__(name, sources=sources, sinks=sinks)

    def dump(self):
        x = super().dump()
        x['flatten'] = {}
        return x


class CodeStep(Step):
    def __init__(self, name=None, source=None, code=None, runtime=None, sources=None, sinks=None, terminator=False):
        super().__init__(name, sources=sources, sinks=sinks, terminator=terminator)
        if source:
            self._source = inspect.getsource(source)
        else:
            self._source = code
        if runtime:
            self._runtime = runtime
        else:
            self._runtime = DEFAULT_RUNTIME

    def dump(self):
        x = super().dump()
        x['code'] = {
            'runtime': self._runtime,
            'source': self._source,
        }
        return x


class MapStep(Step):
    def __init__(self, name=None, expression=None, sources=None, sinks=None):
        super().__init__(name, sources=sources, sinks=sinks)
        assert expression
        self._expression = expression

    def dump(self):
        x = super().dump()
        x['map'] = {
            'expression': self._expression
        }
        return x


class Source:
    def __init__(self, name=None, retry=None):
        self._name = name
        self._retry = retry

    def dump(self):
        x = {}
        if self._name:
            x['name'] = self._name
        if self._retry:
            x['retry'] = self._retry
        return x

    def cat(self, name=None):
        return CatStep(name, sources=[self])

    def container(self, name=None, image=None, args=None, fifo=False, volumes=None, volumeMounts=None, env=None,
                  resources=None,
                  terminator=False):
        return ContainerStep(name, sources=[self], image=image, args=args, fifo=fifo, volumes=volumes,
                             volumeMounts=volumeMounts, env=env, resources=resources, terminator=terminator)

    def dedupe(self, name=None):
        return DedupeStep(name, sources=[self])

    def expand(self, name=None):
        return ExpandStep(name, sources=[self])

    def filter(self, name=None, expression=None):
        return FilterStep(name, expression, sources=[self])

    def git(self, name=None, url=None, branch=None, path=None, image=None, env=None, command=None):
        return GitStep(name, url, branch, path, image, sources=[self], env=env, command=command)

    def group(self, name=None, key=None, format=None, endOfGroup=None, storage=None):
        return GroupStep(name, key, format, endOfGroup, storage, sources=[self])

    def flatten(self, name=None):
        return FlattenStep(name, sources=[self])

    def code(self, name=None, source=None, code=None, runtime=None):
        return CodeStep(name, source=source, code=code, runtime=runtime, sources=[self])

    def map(self, name=None, expression=None):
        return MapStep(name, expression, sources=[self])


def cat(name=None):
    return CatStep(name)


def container(name=None, image=None, args=None, fifo=False, volumes=None, volumeMounts=None, env=None, resources=None,
              terminator=False):
    return ContainerStep(name, terminator=terminator, image=image, args=args, fifo=fifo, volumes=volumes,
                         volumeMounts=volumeMounts, env=env, resources=resources)


def dedupe(name=None):
    return DedupeStep(name)


def expand(name=None):
    return ExpandStep(name)


def filter(name=None, filter=None):
    return FilterStep(name, filter)


def git(name=None, url=None, branch=None, path=None, image=None, env=None, command=None):
    return GitStep(name, url, branch, path, image, env=env, command=command)


def group(name=None, key=None, format=None, endOfGroup=None, storage=None):
    return GroupStep(name, key, format, endOfGroup, storage)


def flatten(name=None):
    return FlattenStep(name)


def handler(name=None, handler=None, code=None, runtime=None):
    return CodeStep(name, handler, code, runtime)


def map(name=None, map=None):
    return MapStep(name, map)


class CronSource(Source):
    def __init__(self, schedule=None, layout=None, name=None, retry=None):
        super().__init__(name=name, retry=retry)
        assert schedule
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
    def __init__(self, name=None, retry=None, serviceName=None):
        super().__init__(name=name, retry=retry)
        self._serviceName = serviceName

    def dump(self):
        x = super().dump()
        h = {}
        if self._serviceName:
            h['serviceName'] = self._serviceName
        x['http'] = h
        return x


class KafkaSource(Source):
    def __init__(self, topic, name=None, retry=None, startOffset=None, fetchMin=None, fetchWaitMax=None):
        super().__init__(name=name, retry=retry)
        assert topic
        self._topic = topic
        self._startOffset = startOffset
        self._fetchMin = fetchMin
        self._fetchWaitMax = fetchWaitMax

    def dump(self):
        x = super().dump()
        y = {'topic': self._topic}
        if self._startOffset:
            y["startOffset"] = self._startOffset
        if self._fetchMin:
            y["fetchMin"] = self._fetchMin
        if self._fetchWaitMax:
            y["fetchWaitMax"] = self._fetchWaitMax
        x['kafka'] = y
        return x


class STANSource(Source):
    def __init__(self, subject, name=None, retry=None):
        super().__init__(name=name, retry=retry)
        assert subject
        self._subject = subject

    def dump(self):
        x = super().dump()
        y = {'subject': self._subject}
        x['stan'] = y
        return x


class JetStreamSource(Source):
    def __init__(self, subject, name=None, retry=None):
        super().__init__(name=name, retry=retry)
        assert subject
        self._subject = subject

    def dump(self):
        x = super().dump()
        y = {'subject': self._subject}
        x['jetstream'] = y
        return x


def cron(schedule=None, layout=None, name=None, retry=None):
    return CronSource(schedule, layout=layout, name=name, retry=retry)


def http(name=None, retry=None, serviceName=None):
    return HTTPSource(name=name, serviceName=serviceName, retry=retry)


def kafka(topic=None, name=None, retry=None, startOffset=None, fetchMin=None, fetchWaitMax=None):
    return KafkaSource(topic, name=name, retry=retry, startOffset=startOffset, fetchMin=fetchMin, fetchWaitMax=fetchWaitMax)


def stan(subject=None, name=None, retry=None):
    return STANSource(subject, name=name, retry=retry)


def jetstream(subject=None, name=None, retry=None):
    return JetStreamSource(subject, name, retry=retry)

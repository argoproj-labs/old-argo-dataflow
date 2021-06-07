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
        self.name = name
        self.annotations = {}
        self.steps = []

    def annotate(self, name, value):
        self.annotations[name] = value
        return self

    def describe(self, value):
        return self.annotate('dataflow.argoproj.io/description', value)

    def step(self, step):
        self.steps.append(step)
        return self

    def build(self):
        return {
            'apiVersion': 'dataflow.argoproj.io/v1alpha1',
            'kind': 'Pipeline',
            'metadata': {
                'name': self.name,
                'annotations': self.annotations
            },
            'spec': {
                'steps': [x.build() for x in self.steps]
            }
        }

    def dump(self):
        sys.stdout.write((yaml.dump(self.build())))


def pipeline(name):
    return PipelineBuilder(name)


class LogSink:
    def build(self):
        return {'log': {}}


class KafkaSink:
    def __init__(self, subject):
        self.subject = subject

    def build(self):
        return {'kafka': {'topic': self.subject}}


class STANSink:
    def __init__(self, topic):
        self.topic = topic

    def build(self):
        return {'stan': {'subject': self.topic}}


class Step:
    def __init__(self, name, sources):
        self.name = name
        self.sources = sources
        self.sinks = []

    def log(self):
        self.sinks.append(LogSink())
        return self

    def kafka(self, subject):
        self.sinks.append(KafkaSink(subject))
        return self

    def stan(self, topic):
        self.sinks.append(STANSink(topic))
        return self

    def build(self):
        return {
            'name': self.name,
            'sources': [x.build() for x in self.sources],
            'sinks': [x.build() for x in self.sinks],
        }


class CatStep(Step):
    def __init__(self, name, sources):
        super().__init__(name, sources)

    def build(self):
        x = super().build()
        x['cat'] = {}
        return x


class ExpandStep(Step):
    def __init__(self, name, sources):
        super().__init__(name, sources)

    def build(self):
        x = super().build()
        x['expand'] = {}
        return x


class FilterStep(Step):
    def __init__(self, name, sources, filter):
        super().__init__(name, sources)
        self.filter = filter

    def build(self):
        x = super().build()
        x['filter'] = self.filter
        return x


class GitStep(Step):
    def __init__(self, name, sources, url, branch, path, image):
        super().__init__(name, sources)
        self.url = url
        self.branch = branch
        self.path = path
        self.image = image

    def build(self):
        x = super().build()
        x['git'] = {
            'url': self.url,
            'branch': self.branch,
            'path': self.path,
            'image': self.image
        }
        return x


class FlattenStep(Step):
    def __init__(self, name, sources):
        super().__init__(name, sources)

    def build(self):
        x = super().build()
        x['flatten'] = {}
        return x


class HandlerStep(Step):
    def __init__(self, name, sources, handler):
        super().__init__(name, sources)
        self.handler = handler

    def build(self):
        x = super().build()
        x['handler'] = {
            'runtime': 'python3-9',
            'code': inspect.getsource(self.handler)
        }
        return x


class MapStep(Step):
    def __init__(self, name, sources, map):
        super().__init__(name, sources)
        self.map = map

    def build(self):
        x = super().build()
        x['map'] = self.map
        return x


class Source:
    def cat(self, name):
        return CatStep(name, [self])

    def expand(self, name):
        return ExpandStep(name, [self])

    def filter(self, name, filter):
        return FilterStep(name, [self], filter)

    def git(self, name, url, branch, path, image):
        return GitStep(name, [self], url, branch, path, image)

    def flatten(self, name):
        return FlattenStep(name, [self])

    def handler(self, name, handler):
        return HandlerStep(name, [self], handler)

    def map(self, name, map):
        return MapStep(name, [self], map)


class CronSource(Source):
    def __init__(self, schedule, layout):
        self.schedule = schedule
        self.layout = layout

    def build(self):
        x = {'schedule': self.schedule}
        if self.layout:
            x['layout'] = self.layout
        return {'cron': x}


class KafkaSource(Source):
    def __init__(self, topic):
        self.topic = topic

    def build(self):
        return {'kafka': {'topic': self.topic}}


class STANSource(Source):
    def __init__(self, subject):
        self.subject = subject

    def build(self):
        return {'stan': {'subject': self.subject}}


def cron(schedule, layout=''):
    return CronSource(schedule, layout)


def kafka(topic):
    return KafkaSource(topic)


def stan(subject):
    return STANSource(subject)

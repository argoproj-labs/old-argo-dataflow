import yaml


class PipelineBuilder():
    def __init__(self, name):
        self.name = name
        self.annotations = {}
        self.steps = []

    def annotate(self, name, value):
        self.annotations[name] = value
        return self

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
                'steps': [x.build(i) for i, x in enumerate(self.steps)]
            }
        }

    def dump(self):
        print(yaml.dump(self.build()))


def pipeline(name):
    return PipelineBuilder(name)


class LogSink:
    def build(self):
        return {'log': {}}


class CatStep():
    def __init__(self, sources):
        self.sources = sources
        self.sinks = []

    def log(self):
        self.sinks.append(LogSink())
        return self

    def build(self, i):
        return {
            'name': 's{i}'.format(i=i),
            'sources': [x.build() for x in self.sources],
            'sinks': [x.build() for x in self.sinks],
            'cat': {}
        }


class CronSource:
    def __init__(self, schedule):
        self.schedule = schedule

    def cat(self):
        return CatStep([self])

    def build(self):
        return {
            'cron': {
                'schedule': self.schedule
            }
        }


def cron(schedule):
    return CronSource(schedule)

from distutils.core import setup

setup(
    name='argo_dataflow_sdk',
    packages=['argo_dataflow_sdk'],
    install_requires=['aiohttp>=3'],
    version='v0.0.84',
    license='apache-2.0',
    description='Argo Dataflow SDK. Can be used to fulfill Argo-Dataflow\'s IMAGE CONTRACT: https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/IMAGE_CONTRACT.md',
    author='Dom Deren',
    author_email='dominik.deren@live.com',
    url='https://github.com/argoproj-labs/argo-dataflow',
    keywords=['Argo', 'Kubernetes'],
    classifiers=[
        'Development Status :: 3 - Alpha',
        'Intended Audience :: Developers',
        'Topic :: Software Development :: Build Tools',
        'License :: OSI Approved :: Apache Software License',
        'Programming Language :: Python',
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.6',
        'Programming Language :: Python :: 3.7',
        'Programming Language :: Python :: 3.8',
        'Programming Language :: Python :: 3.9',
    ],
)

from setuptools import setup, find_packages

setup(
    name="detectviz-adk",
    version="0.1.0",
    packages=find_packages(where="src"),
    package_dir={"": "src"},
    install_requires=[
        "google-adk>=0.0.1",
        "litellm",
        "grpcio>=1.66.1",
        "protobuf>=5.27.3",
        "PyYAML>=6.0.2",
        "jsonschema>=4.23.0",
        "opentelemetry-api>=1.25.0",
        "redis",
        "jinja2",
        "matplotlib",
    ],
    python_requires=">=3.8",
)
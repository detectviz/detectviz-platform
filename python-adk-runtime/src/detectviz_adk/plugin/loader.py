import importlib, json, os, jsonschema, yaml

_LOADED = {}

def load_plugin(manifest_path: str):
    doc = yaml.safe_load(open(manifest_path))
    schema_path = os.path.join("..","contracts","schemas","plugin.schema.json")
    if os.path.exists(schema_path):
        schema = json.load(open(schema_path))
        jsonschema.Draft202012Validator(schema).validate(doc)
    entry = doc["entryPoint"]
    mod, _, sym = entry.partition(":")
    module = importlib.import_module(mod)
    obj = getattr(module, sym)
    _LOADED[doc["id"]] = obj
    return doc["id"]

def unload_plugin(plugin_id: str):
    if plugin_id in _LOADED:
        del _LOADED[plugin_id]

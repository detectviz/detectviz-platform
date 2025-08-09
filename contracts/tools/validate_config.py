import sys, json, yaml, jsonschema
jsonschema.Draft202012Validator(json.load(open(sys.argv[2]))).validate(yaml.safe_load(open(sys.argv[1])))
print('config.yaml OK')

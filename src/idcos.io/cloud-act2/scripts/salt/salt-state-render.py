#!/opt/pyenv/bin/python
# coding=utf-8

import jinja2
import sys
import json


data = sys.stdin.read()
data = json.loads(data)
temp = data.pop('template')
context = data.pop('context')

template = jinja2.Template(temp)
result = template.render(context)
print(result)


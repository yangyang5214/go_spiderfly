import json
import os

import requests
from flask import Flask, Response, request
from werkzeug.routing import BaseConverter

app = Flask(__name__)

domain = 'https://10.0.83.1'


class RegexConverter(BaseConverter):

    def __init__(self, map, regex):
        super(RegexConverter, self).__init__(map)
        self.regex = regex


def service_svc(request):
    args = request.args
    if not args:
        return Response(status=404)
    action = args.get('action', '')
    if not action:
        return Response(status=404)
    return read_file('json/{}'.format(action))


def sub_path(original_url=None):
    path = request.path[1:]
    if path == '':
        path = "index.html"
    final_path = os.path.join(os.getcwd(), 'tmp', path)
    if os.path.exists(final_path) and os.path.isfile(final_path):
        return read_file(final_path)

    # use full_path try
    final_path = os.path.join(os.getcwd(), 'tmp', request.full_path[1:])
    if os.path.exists(final_path) and os.path.isfile(final_path):
        return read_file(final_path)

    html_map = {}
    source = html_map.get(request.path, html_map.get(request.full_path))
    if not source:
        print(domain + request.full_path)
        return Response(status=404)

    return read_file(source)


def read_file(path):
    if not os.path.exists(path):
        return Response(status=404)
    with open(path, 'rb') as f:
        content = f.read()
    content_type = 'Content-Type: text/html; charset=utf-8'
    try:
        json.loads(content)
        content_type = 'application/json; charset=UTF-8'
    except:
        pass
    return Response(content, status=200, content_type=content_type)


app.url_map.converters['regex'] = RegexConverter

if __name__ == '__main__':
    port = 8088
    _uri = '/<regex(r".*"):original_url>'.format()
    app.add_url_rule(_uri, view_func=sub_path, methods=["GET", "POST"])

    app.run(port=port, host='0.0.0.0', debug=False, threaded=True)

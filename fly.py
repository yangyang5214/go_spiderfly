import os

from flask import Flask, Response, request
from werkzeug.routing import BaseConverter

app = Flask(__name__)


class RegexConverter(BaseConverter):

    def __init__(self, map, regex):
        super(RegexConverter, self).__init__(map)
        self.regex = regex


root_dir = None
port = None


def get_file_path(path, full_path):
    file_path = os.path.join(root_dir, path)
    if os.path.exists(file_path):
        return file_path
    file_path = os.path.join(root_dir, full_path)
    if os.path.exists(file_path):
        return file_path

    # 使用 extra 拼接
    file_path = os.path.join(root_dir, path.rstrip('/') + '.extra')
    if os.path.exists(file_path):
        return file_path
    return


def sub_path(original_url=None):
    path = request.path
    if path == '/':
        path = "/index.html"
    final_path = get_file_path(path[1:], request.full_path[1:])
    if final_path:
        return read_file(final_path)
    return Response(status=404)


def read_file(path):
    if os.path.isdir(path):
        path = path[:-1] + '.extra'
    if not os.path.exists(path):
        return Response()
    with open(path, 'rb') as f:
        content = f.read()
    content_type = None
    file_basename = os.path.basename(path).split('?')[0]
    if file_basename.endswith('.css'):
        content_type = 'text/css'
    elif file_basename.endswith('.js'):
        content_type = 'application/javascript'
    elif file_basename.endswith('.png'):
        content_type = 'image/png'
    elif file_basename.endswith('.jpg'):
        content_type = 'image/jpg'
    elif file_basename.endswith('.jpeg'):
        content_type = 'image/jpeg'
    return Response(content, status=200, content_type=content_type)


app.url_map.converters['regex'] = RegexConverter

if __name__ == '__main__':
    import sys

    try:
        root_dir = sys.argv[1]
        port = sys.argv[2]
    except:
        print("python3 fly.py ./tmp/10.0.83.35 8080")
        exit(-1)

    _uri = '/<regex(r".*"):original_url>'.format()
    app.add_url_rule(_uri, view_func=sub_path, methods=["GET", "POST"])

    _uri = '/static/<regex(r".*"):original_url>'.format()
    app.add_url_rule(_uri, view_func=sub_path, methods=["GET", "POST"])

    app.run(port=port, host='0.0.0.0', debug=False, threaded=True)

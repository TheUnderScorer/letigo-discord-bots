import datetime
from functools import update_wrapper

from Cython import basestring
from flask import Flask, current_app, request, make_response
from gpt4all import GPT4All

app = Flask(__name__)

model = GPT4All("Meta-Llama-3-8B-Instruct.Q4_0.gguf")


def cross_domain(origin=None, methods=None, headers=None,
                 max_age=21600, attach_to_all=True,
                 automatic_options=True):
    if methods is not None:
        methods = ', '.join(sorted(x.upper() for x in methods))
    if headers is not None and not isinstance(headers, basestring):
        headers = ', '.join(x.upper() for x in headers)
    if not isinstance(origin, basestring):
        origin = ', '.join(origin)
    if isinstance(max_age, datetime.timedelta):
        max_age = max_age.total_seconds()

    def get_methods():
        if methods is not None:
            return methods

        options_resp = current_app.make_default_options_response()
        return options_resp.headers['allow']

    def decorator(f):
        def wrapped_function(*args, **kwargs):
            if automatic_options and request.method == 'OPTIONS':
                resp = current_app.make_default_options_response()
            else:
                resp = make_response(f(*args, **kwargs))
            if not attach_to_all and request.method != 'OPTIONS':
                return resp

            h = resp.headers

            h['Access-Control-Allow-Origin'] = origin
            h['Access-Control-Allow-Methods'] = get_methods()
            h['Access-Control-Max-Age'] = str(max_age)
            if headers is not None:
                h['Access-Control-Allow-Headers'] = headers
            return resp

        f.provide_automatic_options = False
        return update_wrapper(wrapped_function, f)

    return decorator


@app.route('/status', methods=['GET'])
def status():
    return "ok"


@app.route('/prompt', methods=['POST'])
def prompt():
    try:
        phrase = request.json["phrase"]
        start = datetime.datetime.now()
        reply = model.generate(phrase, max_tokens=2048)
        end = datetime.datetime.now()
        duration = end - start

        return {
            "reply": reply,
            "duration": duration.total_seconds()
        }
    except Exception as e:
        return {
            "error": str(e)
        }, 400

if __name__ == "__main__":
    from waitress import serve
    print('serving via waitress')
    serve(app, host="0.0.0.0", port=8082)
import datetime
import os
import threading
import uuid
from functools import update_wrapper

import torch
from Cython import basestring
from TTS.api import TTS
from flask import Flask, request, make_response, send_file, current_app

app = Flask(__name__)

log = app.logger

device = "cuda" if torch.cuda.is_available() else "cpu"

log.info("Loading model...")

tts = TTS("xtts").to(device)

log.info("Model loaded")

log.info(f"using device: {device}")


speakers_map = {
    'tadeusz': ["speakers/tadeusz/1.wav", "speakers/tadeusz/2.wav", "speakers/tadeusz/3.wav"],
}

files_map = {}


def generate_rand_wav_file():
    return f"{uuid.uuid4()}.wav"


def cleanup_interval():
    log.info(f"doing cleanup. got {len(files_map)} files")

    for sentence in files_map:
        if files_map[sentence]["delete_at"] < datetime.datetime.now():
            log.info(f"removing {sentence} file {files_map[sentence]['file_name']}")
            os.remove(files_map[sentence]["file_name"])
            del files_map[sentence]

    threading.Timer(30.0, cleanup_interval).start()


cleanup_interval()


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


@app.route("/status", methods=["GET"])
def status():
    return "ok"

@app.route("/generate", methods=["POST", "GET"])
@cross_domain(origin="*")
def generate():
    error = None

    sentence = request.json["sentence"]
    speaker = request.json["speaker"]

    log.info(f"got sentence {sentence} and speaker {speaker}")

    if speaker not in speakers_map:
        error = "Speaker not found"

    if sentence == "":
        error = "Empty sentence"

    if error is not None:
        log.error(f"got error {error} for sentence {sentence} and speaker {speaker}")

        return error, 400

    if sentence in files_map:
        return send_file(files_map[sentence]["file_name"], mimetype="audio/wav")

    file_name = generate_rand_wav_file()
    path = f"out/{file_name}"

    tts.tts_to_file(
        text=sentence,
        speaker_wav=speakers_map[speaker], file_path=path, split_sentences=False, language="pl")

    log.info(f"generated file {file_name} for sentence {sentence} and speaker {speaker}")

    files_map[sentence] = {
        "file_name": path,
        "delete_at": datetime.datetime.now() + datetime.timedelta(minutes=30)
    }

    return send_file(path, mimetype="audio/wav")

if __name__ == "__main__":
    from waitress import serve
    serve(app, host="0.0.0.0", port=8080)
